// Package engine provides the chat business logic, independent of any UI.
// It orchestrates LangRails (LLM), GuardRails (safety), and MemoryRails (memory).
package engine

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/promptrails/guardrails"
	"github.com/promptrails/guardrails/scanners"
	"github.com/promptrails/langrails"
	"github.com/promptrails/langrails/openai"
	"github.com/promptrails/memoryrails"
	oaiEmbed "github.com/promptrails/memoryrails/embedders/openai"
	"github.com/promptrails/memoryrails/stores/inmemory"
	"go.uber.org/zap"

	"github.com/promptrails/go-ai-toolkit/internal/config"
)

// Chat implements Engine using LangRails + GuardRails + MemoryRails.
type Chat struct {
	provider langrails.Provider
	memory   *memoryrails.Manager
	guard    *guardrails.Guard
	history  *History
	logger   *zap.Logger
	model    string
	memOn    bool
}

// NewChat creates a new chat engine from config. It initializes the LLM provider,
// guardrails scanners, memory manager, and chat history.
func NewChat(cfg *config.Config, db *sql.DB, logger *zap.Logger) (*Chat, error) {
	history, err := NewHistory(db)
	if err != nil {
		return nil, fmt.Errorf("failed to init history: %w", err)
	}

	// LangRails: unified LLM provider
	provider := openai.New(cfg.APIKey)
	logger.Info("LLM provider initialized", zap.String("model", cfg.Model))

	// GuardRails: content safety scanners
	guard := guardrails.New(
		guardrails.WithScanner(scanners.NewPromptInjection(), guardrails.ActionBlock),
		guardrails.WithScanner(scanners.NewPII(), guardrails.ActionRedact),
		guardrails.WithScanner(scanners.NewSecrets(), guardrails.ActionRedact),
	)
	logger.Info("guardrails initialized", zap.Int("scanners", 3))

	// MemoryRails: semantic agent memory
	var memMgr *memoryrails.Manager
	if cfg.Memory {
		embedder := oaiEmbed.New(cfg.APIKey)
		store := inmemory.New()
		memMgr = memoryrails.NewManager(embedder, store)
		logger.Info("memory initialized", zap.String("embedder", "openai"), zap.String("store", "inmemory"))
	}

	return &Chat{
		provider: provider,
		memory:   memMgr,
		guard:    guard,
		history:  history,
		logger:   logger,
		model:    cfg.Model,
		memOn:    cfg.Memory,
	}, nil
}

// Send processes a user message through the full pipeline:
// guardrails → memory recall → LLM → guardrails → persist.
func (c *Chat) Send(ctx context.Context, userInput string) (string, error) {
	// GuardRails: scan input
	inputResult := c.guard.Scan(ctx, userInput)
	if !inputResult.Passed {
		c.logger.Warn("input blocked by guardrails", zap.String("reason", inputResult.Reason()))
		return "", fmt.Errorf("input blocked: %s", inputResult.Reason())
	}
	safeInput := inputResult.Content
	if inputResult.Redacted {
		c.logger.Info("input redacted by guardrails")
	}

	// Build messages with history and memory context
	msgs, err := c.buildMessages(ctx, safeInput)
	if err != nil {
		return "", err
	}

	// LangRails: call LLM
	c.logger.Debug("calling LLM", zap.String("model", c.model), zap.Int("messages", len(msgs)))
	resp, err := c.provider.Complete(ctx, &langrails.CompletionRequest{
		Model:        c.model,
		SystemPrompt: c.systemPrompt(),
		Messages:     msgs,
	})
	if err != nil {
		c.logger.Error("LLM call failed", zap.Error(err))
		return "", fmt.Errorf("LLM error: %w", err)
	}
	c.logger.Debug("LLM response received",
		zap.Int("prompt_tokens", resp.Usage.PromptTokens),
		zap.Int("completion_tokens", resp.Usage.CompletionTokens))

	// GuardRails: scan output
	outputResult := c.guard.Scan(ctx, resp.Content)
	safeOutput := outputResult.Content
	if outputResult.Redacted {
		c.logger.Info("output redacted by guardrails")
	}

	// Persist to chat history
	_ = c.history.Add("user", safeInput)
	_ = c.history.Add("assistant", safeOutput)

	// MemoryRails: store conversation for future recall
	if c.memOn && c.memory != nil {
		go func() {
			if _, err := c.memory.Remember(context.Background(), safeInput, memoryrails.TypeConversation, nil); err != nil {
				c.logger.Warn("failed to store memory", zap.Error(err))
			}
		}()
	}

	return safeOutput, nil
}

// Execute runs a slash command and returns the result.
func (c *Chat) Execute(ctx context.Context, cmd Command) CommandResult {
	c.logger.Debug("executing command", zap.String("command", string(cmd)))

	switch cmd {
	case CmdClear:
		_ = c.history.Clear()
		return CommandResult{Response: "Chat history cleared."}

	case CmdMemory:
		if !c.memOn || c.memory == nil {
			return CommandResult{Response: "Memory is disabled. Set AI_MEMORY=true to enable."}
		}
		memories, err := c.memory.List(ctx, memoryrails.ListOptions{Limit: 10, Descending: true})
		if err != nil || len(memories) == 0 {
			return CommandResult{Response: "No memories stored yet."}
		}
		var sb strings.Builder
		sb.WriteString("Recent memories:\n")
		for i, m := range memories {
			sb.WriteString(fmt.Sprintf("  %d. [%s] %s\n", i+1, m.Type, truncate(m.Content, 80)))
		}
		return CommandResult{Response: sb.String()}

	case CmdForget:
		if !c.memOn || c.memory == nil {
			return CommandResult{Response: "Memory is disabled."}
		}
		memories, _ := c.memory.List(ctx, memoryrails.ListOptions{})
		for _, m := range memories {
			_ = c.memory.Forget(ctx, m.ID)
		}
		c.logger.Info("all memories forgotten")
		return CommandResult{Response: "All memories forgotten."}

	case CmdHelp:
		var sb strings.Builder
		sb.WriteString("Commands:\n")
		for cmd, desc := range AllCommands() {
			sb.WriteString(fmt.Sprintf("  %-10s %s\n", cmd, desc))
		}
		return CommandResult{Response: sb.String()}

	case CmdQuit, CmdExit:
		return CommandResult{Quit: true}

	default:
		return CommandResult{Response: fmt.Sprintf("Unknown command: %s. Type /help for available commands.", cmd)}
	}
}

// IsCommand returns true if the input starts with /.
func (c *Chat) IsCommand(input string) bool {
	return strings.HasPrefix(input, "/")
}

// ParseCommand extracts the Command enum from user input.
func (c *Chat) ParseCommand(input string) Command {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return ""
	}
	return Command(parts[0])
}

// Model returns the configured LLM model name.
func (c *Chat) Model() string {
	return c.model
}

// buildMessages constructs the message list from history and memory context.
func (c *Chat) buildMessages(ctx context.Context, currentInput string) ([]langrails.Message, error) {
	var msgs []langrails.Message

	// Load recent chat history
	history, _ := c.history.Load(20)
	for _, h := range history {
		msgs = append(msgs, langrails.Message{Role: h.Role, Content: h.Content})
	}

	// MemoryRails: inject relevant memories
	if c.memOn && c.memory != nil {
		results, err := c.memory.Recall(ctx, currentInput, memoryrails.RecallOptions{
			Limit:     5,
			Threshold: 0.3,
		})
		if err == nil && len(results) > 0 {
			c.logger.Debug("recalled memories", zap.Int("count", len(results)))
			var memContext strings.Builder
			memContext.WriteString("[Relevant memories from past conversations]\n")
			for _, r := range results {
				memContext.WriteString(fmt.Sprintf("- %s\n", r.Memory.Content))
			}
			msgs = append([]langrails.Message{
				{Role: "user", Content: memContext.String()},
				{Role: "assistant", Content: "I'll keep this context in mind."},
			}, msgs...)
		}
	}

	msgs = append(msgs, langrails.Message{Role: "user", Content: currentInput})
	return msgs, nil
}

func (c *Chat) systemPrompt() string {
	prompt := "You are a helpful AI assistant. Be concise and clear."
	if c.memOn {
		prompt += " You have access to memories from past conversations and can recall relevant context."
	}
	return prompt
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
