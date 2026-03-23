# Memory & RAG

The toolkit uses [MemoryRails](https://github.com/promptrails/memoryrails) to give the LLM context from past conversations. This is a simple form of RAG (Retrieval-Augmented Generation) — relevant memories are recalled via semantic search and injected into the prompt.

## How It Works

```
User: "What did we discuss about the API redesign?"
  │
  ├─→ MemoryRails: embed the query → search vector store
  │   └─→ Found 3 memories (similarity > 0.3)
  │
  ├─→ Inject memories into prompt as context
  │   [Relevant memories from past conversations]
  │   - We agreed to use REST instead of GraphQL
  │   - The deadline is March 15th
  │   - Auth will use JWT with refresh tokens
  │
  ├─→ LLM sees the memories + current message → crafts informed response
  │
  └─→ Current message stored as new memory for future recall
```

## Configuration

Memory is enabled by default. Disable it with:

```bash
export AI_MEMORY=false
```

## Memory Lifecycle

### Remember (Store)

After each conversation turn, the user's message is stored as a memory:

```go
mem, err := mgr.Remember(ctx, content, memoryrails.TypeConversation, nil)
// mem.ID        = unique identifier
// mem.Content   = the stored text
// mem.Type      = "conversation"
// mem.Embedding = vector representation
// mem.Importance = 0.5 (default)
```

### Recall (Search)

Before each LLM call, relevant memories are retrieved via semantic search:

```go
results, err := mgr.Recall(ctx, query, memoryrails.RecallOptions{
    Limit:     5,     // max memories to return
    Threshold: 0.3,   // minimum similarity score (0-1)
})

for _, r := range results {
    fmt.Printf("%.2f: %s\n", r.Similarity, r.Memory.Content)
}
```

### Forget

Remove specific memories or clear all:

```go
// Forget one memory
mgr.Forget(ctx, memoryID)

// Forget all (via /forget command in the demo app)
memories, _ := mgr.List(ctx, memoryrails.ListOptions{})
for _, m := range memories {
    mgr.Forget(ctx, m.ID)
}
```

## Memory Types

MemoryRails supports different memory categories:

```go
memoryrails.TypeConversation  // chat messages (used by demo app)
memoryrails.TypeFact          // extracted facts
memoryrails.TypeEpisodic      // experiences/events
memoryrails.TypeProcedural    // how-to knowledge
```

The demo app stores everything as `TypeConversation`. In your own app, you can categorize for better retrieval:

```go
// Store a fact extracted from conversation
mgr.Remember(ctx, "User prefers dark mode", memoryrails.TypeFact, nil)

// Store a procedure
mgr.Remember(ctx, "To deploy: run make build && make push", memoryrails.TypeProcedural, nil)
```

## Embedders

MemoryRails needs an embedder to convert text into vectors. The demo app uses OpenAI:

```go
import oaiEmbed "github.com/promptrails/memoryrails/embedders/openai"

embedder := oaiEmbed.New(apiKey)
```

Available embedders:

| Embedder | Package | Requires |
|----------|---------|----------|
| OpenAI | `memoryrails/embedders/openai` | `OPENAI_API_KEY` |
| Ollama | `memoryrails/embedders/ollama` | Local Ollama instance |
| Cohere | `memoryrails/embedders/cohere` | Cohere API key |
| Voyage AI | `memoryrails/embedders/voyage` | Voyage API key |
| Mock | `memoryrails/embedders/mock` | Nothing (for testing) |

## Vector Stores

Embeddings are stored in a vector store for similarity search. The demo app uses in-memory:

```go
import "github.com/promptrails/memoryrails/stores/inmemory"

store := inmemory.New()
```

Available stores:

| Store | Package | Persistence | Best For |
|-------|---------|-------------|----------|
| In-Memory | `stores/inmemory` | None (lost on restart) | Demo, testing |
| SQLite | `stores/sqlite` | File | Single-user apps |
| pgvector | `stores/pgvector` | PostgreSQL | Production |
| Qdrant | `stores/qdrant` | Qdrant server | Large-scale |

### Using SQLite for Persistence

The demo app uses in-memory storage, so memories are lost on restart. Switch to SQLite for persistence:

```go
import "github.com/promptrails/memoryrails/stores/sqlite"

store, err := sqlite.New(filepath.Join(dataDir, "memories.db"))
mgr := memoryrails.NewManager(embedder, store)
```

### Using pgvector for Production

```go
import "github.com/promptrails/memoryrails/stores/pgvector"

store, err := pgvector.New(pgvector.Config{
    DSN:       "postgres://user:pass@localhost/mydb",
    TableName: "memories",
    Dimension: 1536, // OpenAI embedding dimension
})
mgr := memoryrails.NewManager(embedder, store)
```

## Tuning Recall

### Similarity Threshold

The threshold controls how relevant a memory must be to be included. Lower = more results but less relevant:

```go
// Strict: only very relevant memories (good for factual recall)
memoryrails.RecallOptions{Threshold: 0.7, Limit: 3}

// Loose: broader context (good for creative tasks)
memoryrails.RecallOptions{Threshold: 0.2, Limit: 10}

// Demo app default
memoryrails.RecallOptions{Threshold: 0.3, Limit: 5}
```

### Importance Scoring

Memories have an importance score (0-1) that decays over time. High-importance memories are prioritized:

```go
// Store with custom importance
mgr.Remember(ctx, content, memoryrails.TypeFact, &memoryrails.RememberOptions{
    Importance: 0.9, // high importance, decays slower
})
```

## Building a RAG Pipeline

For document-based RAG (not just conversation memory), combine MemoryRails with document loading:

```go
// 1. Load and chunk documents
chunks := splitDocument(documentText, 500) // 500 chars per chunk

// 2. Store each chunk as a memory
for _, chunk := range chunks {
    mgr.Remember(ctx, chunk, memoryrails.TypeFact, nil)
}

// 3. At query time, recall relevant chunks
results, _ := mgr.Recall(ctx, userQuery, memoryrails.RecallOptions{
    Limit:     5,
    Threshold: 0.4,
})

// 4. Inject into prompt
var context strings.Builder
for _, r := range results {
    context.WriteString(r.Memory.Content + "\n\n")
}

resp, _ := provider.Complete(ctx, &langrails.CompletionRequest{
    Model: "gpt-4o",
    SystemPrompt: "Answer based on the provided context. If the context doesn't contain the answer, say so.",
    Messages: []langrails.Message{
        {Role: "user", Content: fmt.Sprintf("Context:\n%s\nQuestion: %s", context.String(), userQuery)},
    },
})
```

## Commands

The demo app includes memory commands:

| Command | Description |
|---------|-------------|
| `/memory` | List recent memories (last 10) |
| `/forget` | Clear all stored memories |
