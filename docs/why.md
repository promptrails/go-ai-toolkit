# Why Go AI Toolkit

## The Problem

Building AI applications in Go means choosing between:

- **Full frameworks** (LangChainGo, Eino) — heavy, opinionated, tightly coupled. You pull in the whole framework even if you just need one feature.
- **Single-purpose libraries** (go-openai, go-anthropic) — provider-locked. Switching means rewriting your integration code.
- **Rolling your own** — HTTP clients for each provider, DIY safety checks, manual memory management. Works until you need the 5th provider.

## The Approach

Go AI Toolkit takes a different path: **small, focused packages** that each do one thing well. You import only what you need.

```
go get github.com/promptrails/langrails     # Need LLM? Get this.
go get github.com/promptrails/guardrails    # Need safety? Get this.
go get github.com/promptrails/memoryrails   # Need memory? Get this.
go get github.com/promptrails/mediarails    # Need media? Get this.
```

Each package:
- Has its own versioning and release cycle
- Works standalone or together
- Follows the same patterns (Provider interface, functional options, zero/minimal deps)
- Is production-tested (extracted from [PromptRails](https://promptrails.com))

## Package Philosophy

| Principle | How |
|-----------|-----|
| **One job per package** | LangRails doesn't do safety. GuardRails doesn't do LLM calls. |
| **Interface-first** | `Provider`, `Scanner`, `Embedder`, `Store` — swap implementations without changing code. |
| **Provider-agnostic** | Switch from OpenAI to Anthropic, or Qdrant to pgvector, by changing one line. |
| **Go-idiomatic** | Channels for streaming, functional options, `context.Context` everywhere. |
| **No magic** | No dependency injection, no reflection, no code generation. Just Go. |

## When to Use What

| You need to... | Use |
|----------------|-----|
| Call any LLM with one interface | [LangRails](https://github.com/promptrails/langrails) |
| Build agent workflows (chains, graphs) | [LangRails](https://github.com/promptrails/langrails) `chain/`, `graph/` |
| Connect to MCP servers or A2A agents | [LangRails](https://github.com/promptrails/langrails) `mcp/`, `a2a/` |
| Scan input for prompt injection | [GuardRails](https://github.com/promptrails/guardrails) |
| Redact PII from LLM output | [GuardRails](https://github.com/promptrails/guardrails) |
| Detect secrets in text | [GuardRails](https://github.com/promptrails/guardrails) |
| Remember facts across conversations | [MemoryRails](https://github.com/promptrails/memoryrails) |
| Semantic search over past context | [MemoryRails](https://github.com/promptrails/memoryrails) |
| Generate speech, images, or video | [MediaRails](https://github.com/promptrails/mediarails) |
| All of the above, working together | [Go AI Toolkit](https://github.com/promptrails/go-ai-toolkit) (this repo) |

## Compared to Alternatives

| | Go AI Toolkit | LangChainGo | Eino | Roll Your Own |
|---|---|---|---|---|
| Import what you need | Yes (4 packages) | No (monolith) | No (monolith) | Yes |
| Provider switching | One line | One line | One line | Rewrite |
| Built-in safety | GuardRails (14 scanners) | No | No | DIY |
| Agent memory | MemoryRails (4 stores) | ConversationBuffer only | Vector store | DIY |
| Media generation | MediaRails (10 providers) | No | No | DIY |
| MCP / A2A protocols | LangRails | No | No | DIY |
| Framework lock-in | None | High | High | None |
