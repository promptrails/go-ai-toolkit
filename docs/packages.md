# Packages Overview

## LangRails — Talk to any LLM

`go get github.com/promptrails/langrails`

The core LLM layer. One `Provider` interface, 12 providers. Swap providers by changing one line.

```go
provider := openai.New("sk-...")          // or anthropic, gemini, ollama, ...
resp, _ := provider.Complete(ctx, &langrails.CompletionRequest{
    Model:    "gpt-4o",
    Messages: []langrails.Message{{Role: "user", Content: "Hello!"}},
})
```

**Includes**: streaming, tool calling, automatic tool loop, prompt templates, memory, structured output, retry/fallback, chain pipelines, graph workflows, MCP client, A2A client+server.

[Documentation](https://promptrails.github.io/langrails) · [GitHub](https://github.com/promptrails/langrails) · [pkg.go.dev](https://pkg.go.dev/github.com/promptrails/langrails)

---

## GuardRails — Keep it safe

`go get github.com/promptrails/guardrails`

Content safety scanning. Scan input before sending to the LLM, scan output before returning to the user.

```go
guard := guardrails.New(
    guardrails.WithScanner(scanners.NewPromptInjection(), guardrails.ActionBlock),
    guardrails.WithScanner(scanners.NewPII(), guardrails.ActionRedact),
)
result := guard.Scan(ctx, userInput)
```

**14 scanners**: PII, toxicity, prompt injection, secrets, ban substrings, invisible text, no refusal, token limit, reading time, JSON validation, URL reachability, ban code, malicious URL, sentiment.

[Documentation](https://promptrails.github.io/guardrails) · [GitHub](https://github.com/promptrails/guardrails) · [pkg.go.dev](https://pkg.go.dev/github.com/promptrails/guardrails)

---

## MemoryRails — Remember everything

`go get github.com/promptrails/memoryrails`

Agent memory with semantic search. Store facts, recall relevant context, let your agent learn from conversations.

```go
mgr := memoryrails.NewManager(embedder, store)
mgr.Remember(ctx, "User prefers dark mode", memoryrails.TypeFact, nil)
results, _ := mgr.Recall(ctx, "user preferences", memoryrails.RecallOptions{Limit: 5})
```

**5 embedders**: OpenAI, Ollama, Cohere, Gemini, Voyage AI
**4 stores**: in-memory, pgvector, SQLite, Qdrant
**5 memory types**: conversation, fact, procedure, episodic, semantic

[Documentation](https://promptrails.github.io/memoryrails) · [GitHub](https://github.com/promptrails/memoryrails) · [pkg.go.dev](https://pkg.go.dev/github.com/promptrails/memoryrails)

---

## MediaRails — Generate media

`go get github.com/promptrails/mediarails`

Speech, image, and video generation through a unified interface. Sync and async providers.

```go
provider := elevenlabs.New("api-key")
resp, _ := provider.Generate(ctx, &mediarails.GenerateRequest{
    Type:   mediarails.TTS,
    Prompt: "Hello world!",
    Config: map[string]any{"voice_id": "21m00Tcm4TlvDq8ikWAM"},
})
```

**10 providers**: OpenAI (TTS/Whisper/DALL-E), ElevenLabs, Deepgram, Stability AI, Fal, Replicate, Runway, Pika, Luma.

[Documentation](https://promptrails.github.io/mediarails) · [GitHub](https://github.com/promptrails/mediarails) · [pkg.go.dev](https://pkg.go.dev/github.com/promptrails/mediarails)

---

## Using Them Together

Each package works standalone. Use one, two, or all four:

```go
// Just LLM
resp, _ := provider.Complete(ctx, req)

// LLM + Safety
input := guard.Scan(ctx, userInput)
resp, _ := provider.Complete(ctx, buildRequest(input.Content))
output := guard.Scan(ctx, resp.Content)

// LLM + Safety + Memory
memories, _ := memory.Recall(ctx, input.Content, recallOpts)
resp, _ := provider.Complete(ctx, buildRequestWithMemory(input.Content, memories))
memory.Remember(ctx, input.Content, memoryrails.TypeConversation, nil)

// All four
input := guard.Scan(ctx, userInput)
memories, _ := memory.Recall(ctx, input.Content, recallOpts)
resp, _ := provider.Complete(ctx, buildRequestWithMemory(input.Content, memories))
output := guard.Scan(ctx, resp.Content)
audio, _ := ttsProvider.Generate(ctx, &mediarails.GenerateRequest{Type: mediarails.TTS, Prompt: output.Content})
```
