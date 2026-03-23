# Error Handling & Retry

Each package in the toolkit has different failure modes. This guide covers how to handle errors and implement retry strategies.

## LLM Errors (LangRails)

LangRails returns `*langrails.APIError` for all HTTP errors from providers:

```go
resp, err := provider.Complete(ctx, req)
if err != nil {
    var apiErr *langrails.APIError
    if errors.As(err, &apiErr) {
        fmt.Printf("Provider: %s\n", apiErr.Provider)
        fmt.Printf("Status: %d\n", apiErr.StatusCode)
        fmt.Printf("Message: %s\n", apiErr.Message)
    }
}
```

### Classifying Errors

```go
var apiErr *langrails.APIError
if errors.As(err, &apiErr) {
    switch {
    case apiErr.IsAuthError():
        // 401/403 — bad API key, don't retry
        log.Fatal("Invalid API key")

    case apiErr.IsRateLimitError():
        // 429 — back off and retry
        time.Sleep(5 * time.Second)

    case apiErr.IsRetryable():
        // 429 or 5xx — safe to retry
        // retry with backoff

    default:
        // 4xx — bad request, don't retry
        log.Printf("Request error: %s", apiErr.Message)
    }
}
```

### Simple Retry with Backoff

```go
func completeWithRetry(ctx context.Context, provider langrails.Provider, req *langrails.CompletionRequest, maxRetries int) (*langrails.CompletionResponse, error) {
    var lastErr error
    for i := 0; i < maxRetries; i++ {
        resp, err := provider.Complete(ctx, req)
        if err == nil {
            return resp, nil
        }
        lastErr = err

        var apiErr *langrails.APIError
        if errors.As(err, &apiErr) && !apiErr.IsRetryable() {
            return nil, err // don't retry 4xx errors
        }

        backoff := time.Duration(1<<uint(i)) * time.Second // 1s, 2s, 4s
        time.Sleep(backoff)
    }
    return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}
```

### Provider Fallback

Switch to a backup provider when the primary fails:

```go
func completeWithFallback(ctx context.Context, req *langrails.CompletionRequest, providers ...langrails.Provider) (*langrails.CompletionResponse, error) {
    for _, p := range providers {
        resp, err := p.Complete(ctx, req)
        if err == nil {
            return resp, nil
        }
        log.Printf("Provider failed, trying next: %v", err)
    }
    return nil, fmt.Errorf("all providers failed")
}

// Usage
primary, _ := llm.New(llm.OpenAI, openaiKey)
fallback, _ := llm.New(llm.Anthropic, anthropicKey)
resp, err := completeWithFallback(ctx, req, primary, fallback)
```

## Media Errors (MediaRails)

Media generation can fail at two stages: the initial request and async polling.

### Sync Providers (DALL-E, ElevenLabs)

```go
resp, err := provider.Generate(ctx, req)
if err != nil {
    // Network error, auth error, or invalid request
    log.Printf("Generation failed: %v", err)
    return err
}
// resp.AssetURL is immediately available
```

### Async Providers (Runway, Pika, Luma, Fal)

Async providers return a job ID. You poll until completion or failure:

```go
resp, err := provider.Generate(ctx, req)
if err != nil {
    return err
}

if resp.JobID != "" {
    // Poll for completion
    for i := 0; i < 30; i++ { // max 30 attempts
        status, err := provider.CheckStatus(ctx, resp.JobID)
        if err != nil {
            return fmt.Errorf("poll failed: %w", err)
        }

        switch status.Status {
        case mediarails.JobCompleted:
            fmt.Println("URL:", status.AssetURL)
            return nil
        case mediarails.JobFailed:
            return fmt.Errorf("generation failed")
        default:
            // still processing
            time.Sleep(5 * time.Second)
        }
    }
    return fmt.Errorf("generation timed out")
}
```

### WaitForCompletion Helper

MediaRails includes a built-in poller with exponential backoff:

```go
final, err := mediarails.WaitForCompletion(ctx, provider, resp.JobID,
    2*time.Second,   // initial interval
    30*time.Second,  // max interval
    10*time.Minute,  // timeout
)
if err != nil {
    log.Printf("Video generation failed: %v", err)
}
```

### Not-Async Error

Some providers are sync-only. Calling `CheckStatus` on them returns `ErrNotAsync`:

```go
status, err := provider.CheckStatus(ctx, jobID)
if errors.Is(err, mediarails.ErrNotAsync) {
    // This provider doesn't support async — result was in Generate() response
}
```

## GuardRails Errors

GuardRails scanners don't return errors — they return results with `Passed: true/false`. A scanner failure (blocked content) is not a Go error:

```go
result := guard.Scan(ctx, input)

// This is NOT an error — it's a policy decision
if !result.Passed {
    // Content was blocked by a scanner
    fmt.Println("Reason:", result.Reason())
    // Decide: reject the request, ask user to rephrase, etc.
}
```

### Scanner vs System Errors

If a scanner itself crashes (e.g., LLM Guard API is down), the behavior depends on the scanner type:

- **Local scanners** (PII, Toxicity, etc.): pure regex — never fail
- **LLM Guard scanners**: API call may fail — the scan is marked as failed with an error message

```go
for _, r := range result.Results {
    if !r.Passed && strings.Contains(r.Message, "request failed") {
        // LLM Guard API error, not a content violation
        log.Printf("Scanner %s unavailable: %s", r.Scanner, r.Message)
    }
}
```

## Tool Calling Errors

When a tool call fails, the error is sent back to the LLM as a tool result. The LLM then decides how to respond:

```go
// Inside tools.RunLoop, errors are automatically wrapped:
// {"error": "city not found: Atlantis"}
// The LLM sees this and can say "I couldn't find that city, try another."
```

### Max Iterations

The tool loop has a safety limit (default: 20, demo app: 5):

```go
result, err := tools.RunLoop(ctx, provider, req, executor,
    tools.WithMaxIterations(5),
)
if err != nil {
    // Could be: "tool loop exceeded maximum iterations (5)"
    // This means the LLM kept calling tools without giving a final answer
}
```

## Context & Timeouts

Always use context for cancellation and timeouts:

```go
// Timeout for the entire operation
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

resp, err := provider.Complete(ctx, req)
if errors.Is(err, context.DeadlineExceeded) {
    log.Println("LLM call timed out")
}
```

For media generation, use longer timeouts:

```go
// Image generation: 60 seconds is usually enough
ctx, cancel := context.WithTimeout(ctx, 60*time.Second)

// Video generation: may take several minutes (async)
ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
```
