# Observability & Tracing

The Go AI Toolkit doesn't include a tracing package — instead, it's designed to work with the Go ecosystem's standard observability tools, primarily [OpenTelemetry](https://opentelemetry.io/).

Each package exposes the data you need to instrument:

## LLM Calls (LangRails)

```go
import "go.opentelemetry.io/otel"

tracer := otel.Tracer("ai-chat")

ctx, span := tracer.Start(ctx, "llm.complete")
defer span.End()

resp, err := provider.Complete(ctx, req)
if err != nil {
    span.RecordError(err)
    span.SetStatus(codes.Error, err.Error())
    return err
}

// Token usage
span.SetAttributes(
    attribute.String("llm.model", req.Model),
    attribute.Int("llm.prompt_tokens", resp.Usage.PromptTokens),
    attribute.Int("llm.completion_tokens", resp.Usage.CompletionTokens),
    attribute.String("llm.finish_reason", resp.FinishReason),
)
```

## Safety Scanning (GuardRails)

```go
ctx, span := tracer.Start(ctx, "guardrails.scan")
defer span.End()

result := guard.Scan(ctx, input)

span.SetAttributes(
    attribute.Bool("guardrails.passed", result.Passed),
    attribute.Bool("guardrails.redacted", result.Redacted),
    attribute.Int("guardrails.scanner_count", len(result.Results)),
)

if !result.Passed {
    span.SetAttributes(attribute.String("guardrails.reason", result.Reason()))
}

// Per-scanner results
for _, r := range result.Results {
    if !r.Passed {
        span.AddEvent("guardrails.violation", trace.WithAttributes(
            attribute.String("scanner", string(r.Scanner)),
            attribute.String("message", r.Message),
        ))
    }
}
```

## Memory Operations (MemoryRails)

```go
// Remember
ctx, span := tracer.Start(ctx, "memory.remember")
mem, err := mgr.Remember(ctx, content, memoryrails.TypeFact, nil)
span.SetAttributes(
    attribute.String("memory.id", mem.ID),
    attribute.String("memory.type", string(mem.Type)),
    attribute.Float64("memory.importance", mem.Importance),
)
span.End()

// Recall
ctx, span = tracer.Start(ctx, "memory.recall")
results, err := mgr.Recall(ctx, query, memoryrails.RecallOptions{Limit: 5})
span.SetAttributes(
    attribute.Int("memory.results", len(results)),
    attribute.String("memory.query", query),
)
if len(results) > 0 {
    span.SetAttributes(
        attribute.Float64("memory.top_similarity", results[0].Similarity),
    )
}
span.End()
```

## Full Request Trace

A typical chat request produces this trace:

```
ai-chat.request
├── guardrails.scan (input)
│   └── scanner: prompt_injection ✓
│   └── scanner: pii → redacted
│   └── scanner: secrets ✓
├── memory.recall
│   └── 3 memories found (top: 0.87)
├── llm.complete
│   └── model: gpt-4o-mini
│   └── tokens: 450 prompt, 120 completion
├── guardrails.scan (output)
│   └── all passed
└── memory.remember (async)
```

## Structured Logging

The demo app uses Zap for structured logging to `~/.ai-chat/ai-chat.log`:

```json
{"level":"info","ts":"2026-03-21T10:00:00Z","msg":"LLM provider initialized","model":"gpt-4o-mini"}
{"level":"debug","ts":"2026-03-21T10:00:05Z","msg":"calling LLM","model":"gpt-4o-mini","messages":5}
{"level":"debug","ts":"2026-03-21T10:00:07Z","msg":"LLM response received","prompt_tokens":450,"completion_tokens":120}
{"level":"info","ts":"2026-03-21T10:00:07Z","msg":"output redacted by guardrails"}
```

## Metrics

Export key metrics via OpenTelemetry or Prometheus:

| Metric | Source | Type |
|--------|--------|------|
| `llm_requests_total` | LangRails | Counter |
| `llm_tokens_total` | `resp.Usage` | Counter |
| `llm_latency_seconds` | Span duration | Histogram |
| `guardrails_blocked_total` | `!result.Passed` | Counter |
| `guardrails_redacted_total` | `result.Redacted` | Counter |
| `memory_recall_count` | `len(results)` | Histogram |
| `memory_top_similarity` | `results[0].Similarity` | Histogram |

## Recommended Setup

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
    "go.opentelemetry.io/otel/sdk/trace"
)

// Export to Jaeger, Grafana Tempo, Datadog, etc.
exporter, _ := otlptracehttp.New(ctx)
tp := trace.NewTracerProvider(trace.WithBatcher(exporter))
otel.SetTracerProvider(tp)
```

Works with: **Jaeger**, **Grafana Tempo**, **Datadog**, **Honeycomb**, **New Relic**, or any OpenTelemetry-compatible backend.
