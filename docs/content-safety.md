# Content Safety

The toolkit uses [GuardRails](https://github.com/promptrails/guardrails) to scan both user input and LLM output for safety violations. Scanning happens automatically on every message — no manual calls needed.

## Default Scanners

The demo app ships with three scanners:

| Scanner | Action | What It Does |
|---------|--------|-------------|
| Prompt Injection | **Block** | Detects jailbreak attempts, instruction overrides, role hijacking |
| PII | **Redact** | Replaces emails, phone numbers, SSNs, credit cards, IPs with labels |
| Secrets | **Redact** | Replaces API keys, tokens, private keys, connection strings |

```go
guard := guardrails.New(
    guardrails.WithScanner(scanners.NewPromptInjection(), guardrails.ActionBlock),
    guardrails.WithScanner(scanners.NewPII(), guardrails.ActionRedact),
    guardrails.WithScanner(scanners.NewSecrets(), guardrails.ActionRedact),
)
```

## Actions

Each scanner is paired with an action that determines what happens on detection:

| Action | Behavior | Use When |
|--------|----------|----------|
| `ActionBlock` | Reject the message entirely | Dangerous content (injection, toxicity) |
| `ActionRedact` | Replace matches, continue processing | Sensitive data (PII, secrets) |
| `ActionLog` | Record the violation, continue unchanged | Monitoring without enforcement |

## Scan Flow

```
User Input
  │
  ├─→ Prompt Injection scanner → BLOCK if detected
  ├─→ PII scanner → REDACT matches (email → [EMAIL], phone → [PHONE])
  ├─→ Secrets scanner → REDACT matches (sk-xxx → [API_KEY])
  │
  ├─→ ... LLM processes the safe/redacted input ...
  │
  └─→ Same scanners run on LLM output before showing to user
```

## Available Scanners

GuardRails includes 14 scanners. The demo uses 3, but you can enable any combination:

```go
guard := guardrails.New(
    // Block dangerous content
    guardrails.WithScanner(scanners.NewPromptInjection(), guardrails.ActionBlock),
    guardrails.WithScanner(scanners.NewToxicity(), guardrails.ActionBlock),

    // Redact sensitive data
    guardrails.WithScanner(scanners.NewPII(), guardrails.ActionRedact),
    guardrails.WithScanner(scanners.NewSecrets(), guardrails.ActionRedact),
    guardrails.WithScanner(scanners.NewBanSubstrings("confidential", "internal"), guardrails.ActionRedact),

    // Monitor without blocking
    guardrails.WithScanner(scanners.NewSentiment(), guardrails.ActionLog),
)
```

### Full Scanner List

| Scanner | Detects | Supports Redaction |
|---------|---------|-------------------|
| `NewPromptInjection()` | Jailbreaks, instruction overrides | No |
| `NewToxicity()` | Offensive language | No |
| `NewPII()` | Email, phone, SSN, credit card, IP | Yes |
| `NewSecrets()` | API keys, tokens, private keys | Yes |
| `NewBanSubstrings(...)` | Custom banned words/phrases | Yes |
| `NewInvisibleText()` | Hidden Unicode characters | Yes |
| `NewNoRefusal()` | LLM refusal phrases ("I'm sorry, I can't...") | No |
| `NewTokenLimit(n)` | Word count over limit | No |
| `NewReadingTime(seconds)` | Content too long to read | No |
| `NewJSONValidator()` | Invalid JSON structure | No |
| `NewURLReachability()` | Hallucinated/dead URLs | No |
| `NewBanCode()` | Code snippets in output | No |
| `NewMaliciousURL()` | Suspicious TLDs, phishing URLs | No |
| `NewSentiment()` | Negative sentiment | No |

## Customizing PII Detection

Filter which PII types to detect:

```go
// Only detect emails and phone numbers
scanner := scanners.NewPIIWithTypes("email", "phone")
```

Available types: `email`, `phone`, `ssn`, `credit_card`, `ip_address`

## Customizing Toxicity

Add domain-specific offensive words:

```go
scanner := scanners.NewToxicityWithWords("spam", "scam", "fraud")
```

## Customizing Prompt Injection

Add custom injection patterns:

```go
scanner := scanners.NewPromptInjectionWithPatterns(
    `(?i)reveal your system prompt`,
    `(?i)what are your instructions`,
)
```

## Checking Results

```go
result := guard.Scan(ctx, userInput)

if !result.Passed {
    // At least one scanner blocked the content
    fmt.Println("Blocked:", result.Reason())
}

if result.Redacted {
    // Content was modified by redaction scanners
    fmt.Println("Original was redacted")
    fmt.Println("Safe content:", result.Content)
}

// Inspect individual scanner results
for _, r := range result.Results {
    fmt.Printf("Scanner: %s, Passed: %v\n", r.Scanner, r.Passed)
    if !r.Passed {
        fmt.Printf("  Message: %s\n", r.Message)
        fmt.Printf("  Matches: %v\n", r.Matches)
    }
}
```

## LLM Guard (ML-Powered)

For higher accuracy scanning using ML models, GuardRails supports [LLM Guard](https://llm-guard.com) as a backend:

```go
import "github.com/promptrails/guardrails/llmguard"

client := llmguard.NewClient("http://localhost:8000", "token")

guard := guardrails.New(
    guardrails.WithScanner(
        llmguard.NewScanner(client, "Toxicity", guardrails.ScannerToxicity),
        guardrails.ActionBlock,
    ),
    guardrails.WithScanner(
        llmguard.NewScanner(client, "Anonymize", guardrails.ScannerPII),
        guardrails.ActionRedact,
    ),
)
```

See the [GuardRails LLM Guard docs](https://promptrails.github.io/guardrails/#/llm-guard) for setup instructions.
