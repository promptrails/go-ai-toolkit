# Go AI Toolkit

Production-ready AI building blocks for Go.

Three packages. One toolkit. Everything you need to build AI-powered applications in Go.

## The Toolkit

| Package | What It Does | |
|---------|-------------|---|
| [**LangRails**](https://github.com/promptrails/langrails) | Unified LLM provider — 12 providers, streaming, tool calling, chain, graph, MCP, A2A | [![Go](https://img.shields.io/github/v/tag/promptrails/langrails)](https://github.com/promptrails/langrails) |
| [**GuardRails**](https://github.com/promptrails/guardrails) | Content safety — PII, toxicity, prompt injection, secrets, sentiment, 14 scanners | [![Go](https://img.shields.io/github/v/tag/promptrails/guardrails)](https://github.com/promptrails/guardrails) |
| [**MemoryRails**](https://github.com/promptrails/memoryrails) | Agent memory — embeddings, vector stores, semantic search, importance decay | [![Go](https://img.shields.io/github/v/tag/promptrails/memoryrails)](https://github.com/promptrails/memoryrails) |
| [**MediaRails**](https://github.com/promptrails/mediarails) | AI media generation — speech, image, video, 10 providers | [![Go](https://img.shields.io/github/v/tag/promptrails/mediarails)](https://github.com/promptrails/mediarails) |

## Demo: AI Chat TUI

This repo includes a terminal chat application that demonstrates all three packages working together.

![AI Chat TUI Demo](docs/demo.gif)

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

**Via Go:**

```bash
go install github.com/promptrails/go-ai-toolkit/cmd/chat@latest
export OPENAI_API_KEY=sk-...
ai-chat
```

**Via GitHub Release:**

Download the latest binary from [Releases](https://github.com/promptrails/go-ai-toolkit/releases), extract, and run:

```bash
export OPENAI_API_KEY=sk-...
./ai-chat
```

**From source:**

```bash
git clone https://github.com/promptrails/go-ai-toolkit
cd go-ai-toolkit
cp .env.example .env  # edit with your API key
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
| MediaRails | [promptrails.github.io/mediarails](https://promptrails.github.io/mediarails) | [pkg.go.dev](https://pkg.go.dev/github.com/promptrails/mediarails) |

## License

MIT — [PromptRails](https://promptrails.com)
