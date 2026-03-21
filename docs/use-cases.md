# Use Cases

Real-world patterns for combining the toolkit packages.

## Customer Support Bot

**Packages**: LangRails + GuardRails + MemoryRails

```go
// Scan user input for safety
input := guard.Scan(ctx, userMessage)
if !input.Passed {
    return "I can't process that request."
}

// Recall past interactions with this customer
memories, _ := memory.Recall(ctx, input.Content, memoryrails.RecallOptions{
    Limit: 5,
    Metadata: map[string]any{"customer_id": customerID},
})

// Build context-aware prompt
resp, _ := provider.Complete(ctx, &langrails.CompletionRequest{
    Model:        "gpt-4o",
    SystemPrompt: "You are a support agent. Use customer history to personalize responses.",
    Messages:     buildMessages(input.Content, memories),
})

// Store interaction for future recall
memory.Remember(ctx, input.Content, memoryrails.TypeConversation,
    map[string]any{"customer_id": customerID})
```

## Content Moderation Pipeline

**Packages**: GuardRails

```go
guard := guardrails.New(
    guardrails.WithScanner(scanners.NewToxicity(), guardrails.ActionBlock),
    guardrails.WithScanner(scanners.NewPII(), guardrails.ActionRedact),
    guardrails.WithScanner(scanners.NewSecrets(), guardrails.ActionRedact),
    guardrails.WithScanner(scanners.NewMaliciousURL(), guardrails.ActionBlock),
    guardrails.WithScanner(scanners.NewBanSubstrings("competitor"), guardrails.ActionLog),
)

// Scan user-generated content
result := guard.Scan(ctx, userContent)
if !result.Passed {
    flagForReview(userContent, result.Reasons())
    return
}
publish(result.Content) // PII and secrets redacted
```

## AI Podcast Generator

**Packages**: LangRails + MediaRails

```go
// Generate script
script, _ := provider.Complete(ctx, &langrails.CompletionRequest{
    Model:        "gpt-4o",
    SystemPrompt: "Write a 2-minute podcast script about the given topic.",
    Messages:     []langrails.Message{{Role: "user", Content: topic}},
})

// Convert to speech
audio, _ := ttsProvider.Generate(ctx, &mediarails.GenerateRequest{
    Type:   mediarails.TTS,
    Model:  "eleven_multilingual_v2",
    Prompt: script.Content,
    Config: map[string]any{"voice_id": "podcast_host_voice"},
})

os.WriteFile("episode.mp3", audio.AssetData, 0644)
```

## Multi-Step Research Agent

**Packages**: LangRails (chain + tools)

```go
// Define tools
executor := tools.NewMap(map[string]tools.Func{
    "search": func(ctx context.Context, args string) (string, error) {
        return searchAPI(args)
    },
    "summarize": func(ctx context.Context, args string) (string, error) {
        resp, _ := provider.Complete(ctx, &langrails.CompletionRequest{
            Model:        "gpt-4o-mini",
            SystemPrompt: "Summarize in 3 bullet points.",
            Messages:     []langrails.Message{{Role: "user", Content: args}},
        })
        return resp.Content, nil
    },
})

// Run tool loop
result, _ := tools.RunLoop(ctx, provider, &langrails.CompletionRequest{
    Model:    "gpt-4o",
    Messages: []langrails.Message{{Role: "user", Content: "Research quantum computing advances in 2025"}},
    Tools:    toolDefinitions,
}, executor)
```

## Image Generation with Safety

**Packages**: GuardRails + MediaRails

```go
// Check prompt for policy violations before generating
input := guard.Scan(ctx, userPrompt)
if !input.Passed {
    return nil, fmt.Errorf("prompt blocked: %s", input.Reason())
}

// Generate image
resp, _ := dalleProvider.Generate(ctx, &mediarails.GenerateRequest{
    Type:   mediarails.ImageGen,
    Model:  "dall-e-3",
    Prompt: input.Content, // safe, redacted prompt
})
```

## Meeting Transcription with Memory

**Packages**: MediaRails + MemoryRails + LangRails

```go
// Transcribe audio
transcript, _ := whisperProvider.Generate(ctx, &mediarails.GenerateRequest{
    Type:      mediarails.STT,
    InputData: audioBytes,
})

// Extract key points
summary, _ := provider.Complete(ctx, &langrails.CompletionRequest{
    Model:        "gpt-4o",
    SystemPrompt: "Extract action items and decisions from this meeting transcript.",
    Messages:     []langrails.Message{{Role: "user", Content: transcript.TextOutput}},
})

// Store as episodic memory
memory.Remember(ctx, summary.Content, memoryrails.TypeEpisodic,
    map[string]any{"meeting_date": time.Now().Format("2006-01-02")})
```

## Stateful Workflow with Approval

**Packages**: LangRails (graph)

```go
g := graph.New[ReviewState]()

g.AddNode("draft", func(ctx context.Context, s ReviewState) (ReviewState, error) {
    resp, _ := provider.Complete(ctx, &langrails.CompletionRequest{
        Model:    "gpt-4o",
        Messages: []langrails.Message{{Role: "user", Content: "Draft: " + s.Topic}},
    })
    s.Draft = resp.Content
    return s, nil
})

g.AddNode("review", func(ctx context.Context, s ReviewState) (ReviewState, error) {
    resp, _ := provider.Complete(ctx, &langrails.CompletionRequest{
        Model:        "gpt-4o",
        SystemPrompt: "Review this draft. Reply APPROVED or list issues.",
        Messages:     []langrails.Message{{Role: "user", Content: s.Draft}},
    })
    s.Review = resp.Content
    return s, nil
})

g.SetEntryPoint("draft")
g.AddEdge("draft", "review")
g.AddConditionalEdge("review", func(s ReviewState) string {
    if strings.Contains(s.Review, "APPROVED") {
        return graph.END
    }
    return "draft" // revise and try again
})

result, _ := g.Run(ctx, ReviewState{Topic: "Q1 report"})
```
