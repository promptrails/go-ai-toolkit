# Architecture

## Directory Structure

```
go-ai-toolkit/
├── cmd/chat/main.go           # Entry point — wires config, engine, TUI
├── internal/
│   ├── config/config.go       # Environment variable configuration
│   ├── engine/                # Business logic (UI-independent)
│   │   ├── engine.go          # Engine interface
│   │   ├── command.go         # Slash command enum
│   │   ├── chat.go            # Engine implementation
│   │   └── history.go         # SQLite chat history
│   └── tui/tui.go             # Bubbletea terminal UI
```

## Data Flow

```
User Input
  │
  ├─→ TUI (internal/tui)
  │     Captures input, renders messages
  │
  ├─→ Engine (internal/engine)
  │     │
  │     ├─→ GuardRails: scan input
  │     │     Block prompt injection
  │     │     Redact PII and secrets
  │     │
  │     ├─→ MemoryRails: recall context
  │     │     Semantic search over past conversations
  │     │     Inject relevant memories into prompt
  │     │
  │     ├─→ LangRails: call LLM
  │     │     Send messages with system prompt + memory
  │     │     Receive response
  │     │
  │     ├─→ GuardRails: scan output
  │     │     Redact any leaked PII/secrets
  │     │
  │     ├─→ History: persist to SQLite
  │     │
  │     └─→ MemoryRails: store for future recall
  │
  └─→ TUI: display response
```

## Packages Used

### LangRails (`github.com/promptrails/langrails`)

Provides the `Provider` interface for LLM communication. In this app, we use the OpenAI provider:

```go
provider := openai.New(cfg.APIKey)
resp, _ := provider.Complete(ctx, &langrails.CompletionRequest{
    Model:    "gpt-4o-mini",
    Messages: messages,
})
```

### GuardRails (`github.com/promptrails/guardrails`)

Scans both input and output for safety:

```go
guard := guardrails.New(
    guardrails.WithScanner(scanners.NewPromptInjection(), guardrails.ActionBlock),
    guardrails.WithScanner(scanners.NewPII(), guardrails.ActionRedact),
    guardrails.WithScanner(scanners.NewSecrets(), guardrails.ActionRedact),
)
result := guard.Scan(ctx, userInput)
```

### MemoryRails (`github.com/promptrails/memoryrails`)

Stores and recalls conversation context:

```go
memory := memoryrails.NewManager(embedder, store)
memory.Remember(ctx, "user preference", memoryrails.TypeFact, nil)
results, _ := memory.Recall(ctx, "query", memoryrails.RecallOptions{Limit: 5})
```

## Engine Interface

The `Engine` interface decouples business logic from UI:

```go
type Engine interface {
    Send(ctx context.Context, input string) (string, error)
    Execute(ctx context.Context, cmd Command) CommandResult
    IsCommand(input string) bool
    ParseCommand(input string) Command
    Model() string
}
```

This makes it possible to:
- Test business logic without TUI
- Swap TUI for a web UI or API
- Mock the engine for UI tests
