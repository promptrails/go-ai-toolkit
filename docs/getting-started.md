# Getting Started

## Install

**Via Go:**

```bash
go install github.com/promptrails/go-ai-toolkit/cmd/chat@latest
```

**Via GitHub Release:**

Download the latest binary from [Releases](https://github.com/promptrails/go-ai-toolkit/releases).

**From source:**

```bash
git clone https://github.com/promptrails/go-ai-toolkit
cd go-ai-toolkit
make build
```

## Configure

```bash
export OPENAI_API_KEY=sk-...
```

Or copy the `.env.example` file:

```bash
cp .env.example .env
# Edit .env with your API key
```

## Run

```bash
./ai-chat
```

Or with `make`:

```bash
make run
```

## Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `OPENAI_API_KEY` | Yes | — | OpenAI API key |
| `AI_MODEL` | No | `gpt-4o-mini` | Model to use |
| `AI_MEMORY` | No | `true` | Enable semantic memory |
| `AI_DATA_DIR` | No | `~/.ai-chat` | Data directory for SQLite |

## What Happens When You Chat

1. Your input goes through **GuardRails** — prompt injection is blocked, PII and secrets are redacted
2. **MemoryRails** recalls relevant memories from past conversations
3. **LangRails** sends your message (with context) to the LLM
4. The response goes through **GuardRails** again — output is scanned and redacted if needed
5. Your message is stored in **MemoryRails** for future recall
6. Both messages are saved to SQLite chat history
