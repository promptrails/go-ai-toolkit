# Go AI Toolkit

Production-ready AI building blocks for Go.

Three packages. One toolkit. Everything you need to build AI-powered applications in Go.

## The Toolkit

| Package | What It Does | |
|---------|-------------|---|
| [**LangRails**](https://github.com/promptrails/langrails) | Unified LLM provider — 12 providers, streaming, tool calling, chain, graph, MCP, A2A | [![Go](https://img.shields.io/github/v/tag/promptrails/langrails)](https://github.com/promptrails/langrails) |
| [**GuardRails**](https://github.com/promptrails/guardrails) | Content safety — PII, toxicity, prompt injection, secrets, sentiment, 14 scanners | [![Go](https://img.shields.io/github/v/tag/promptrails/guardrails)](https://github.com/promptrails/guardrails) |
| [**MemoryRails**](https://github.com/promptrails/memoryrails) | Agent memory — embeddings, vector stores, semantic search, importance decay | [![Go](https://img.shields.io/github/v/tag/promptrails/memoryrails)](https://github.com/promptrails/memoryrails) |

## Demo: AI Chat TUI

This repo includes a terminal chat application that demonstrates all three packages working together.

```
┌─ AI Chat  model:gpt-4o-mini ─────────────────────┐
│                                                    │
│  you: What's the capital of Turkey?                │
│                                                    │
│  ai: The capital of Turkey is Ankara.              │
│                                                    │
│  you: Remember that I'm learning Turkish           │
│                                                    │
│  ai: Got it! I'll keep that in mind.               │
│                                                    │
│  you: /memory                                      │
│                                                    │
│  system: Recent memories:                          │
│    1. [conversation] What's the capital of Turkey?  │
│    2. [conversation] Remember that I'm learning... │
│                                                    │
├────────────────────────────────────────────────────┤
│ Type a message... (/help for commands)             │
└────────────────────────────────────────────────────┘
```

### How It Works

```
User Input
  │
  ├─→ GuardRails: scan for injection, redact PII & secrets
  │
  ├─→ MemoryRails: recall relevant memories from past conversations
  │
  ├─→ LangRails: send to LLM with context + memory
  │
  ├─→ GuardRails: scan output, redact if needed
  │
  └─→ MemoryRails: store conversation for future recall
```

### Try It

```bash
# Install
go install github.com/promptrails/go-ai-toolkit/cmd/chat@latest

# Run
export OPENAI_API_KEY=sk-...
ai-chat
```

Or build from source:

```bash
git clone https://github.com/promptrails/go-ai-toolkit
cd go-ai-toolkit
make run
```

### Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `OPENAI_API_KEY` | Yes | — | OpenAI API key |
| `AI_MODEL` | No | `gpt-4o-mini` | Model to use |
| `AI_MEMORY` | No | `true` | Enable semantic memory |
| `AI_DATA_DIR` | No | `~/.ai-chat` | Data directory for SQLite |

### Commands

| Command | Description |
|---------|-------------|
| `/help` | Show available commands |
| `/memory` | List stored memories |
| `/forget` | Clear all memories |
| `/clear` | Clear chat history |
| `/quit` | Exit |

## Using the Toolkit in Your App

```go
import (
    "github.com/promptrails/langrails"
    "github.com/promptrails/langrails/openai"
    "github.com/promptrails/guardrails"
    "github.com/promptrails/guardrails/scanners"
    "github.com/promptrails/memoryrails"
    oaiEmbed "github.com/promptrails/memoryrails/embedders/openai"
    "github.com/promptrails/memoryrails/stores/inmemory"
)

// 1. LLM Provider
provider := openai.New("sk-...")

// 2. Safety Guard
guard := guardrails.New(
    guardrails.WithScanner(scanners.NewPromptInjection(), guardrails.ActionBlock),
    guardrails.WithScanner(scanners.NewPII(), guardrails.ActionRedact),
)

// 3. Agent Memory
memory := memoryrails.NewManager(
    oaiEmbed.New("sk-..."),
    inmemory.New(),
)

// Wire them together
input := guard.Scan(ctx, userInput)
if !input.Passed { /* blocked */ }

memories, _ := memory.Recall(ctx, input.Content, memoryrails.RecallOptions{Limit: 5})
resp, _ := provider.Complete(ctx, &langrails.CompletionRequest{
    Model:    "gpt-4o",
    Messages: buildMessages(input.Content, memories),
})

output := guard.Scan(ctx, resp.Content)
memory.Remember(ctx, input.Content, memoryrails.TypeConversation, nil)
```

## Documentation

| Package | Docs | Go Reference |
|---------|------|-------------|
| LangRails | [promptrails.github.io/langrails](https://promptrails.github.io/langrails) | [pkg.go.dev](https://pkg.go.dev/github.com/promptrails/langrails) |
| GuardRails | [promptrails.github.io/guardrails](https://promptrails.github.io/guardrails) | [pkg.go.dev](https://pkg.go.dev/github.com/promptrails/guardrails) |
| MemoryRails | [promptrails.github.io/memoryrails](https://promptrails.github.io/memoryrails) | [pkg.go.dev](https://pkg.go.dev/github.com/promptrails/memoryrails) |

## License

MIT — [PromptRails](https://promptrails.com)
