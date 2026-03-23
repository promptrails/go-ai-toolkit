# Custom Tools

LangRails provides a tool calling loop that lets the LLM call functions you define. The toolkit uses this for media generation, but you can add any tool — web search, database queries, API calls, calculations, etc.

## How Tool Calling Works

```
1. You define tools with names, descriptions, and JSON Schema parameters
2. The LLM sees these tools and decides when to call them
3. LangRails executes the tool and sends the result back to the LLM
4. The LLM uses the result to craft its final response
```

## Defining a Tool

A tool has two parts: a **definition** (what the LLM sees) and a **function** (what runs when called).

```go
import (
    "github.com/promptrails/langrails"
    "github.com/promptrails/langrails/tools"
)

// 1. Define the tool function
weatherFunc := func(ctx context.Context, arguments string) (string, error) {
    var args struct {
        City string `json:"city"`
    }
    json.Unmarshal([]byte(arguments), &args)

    // Call your API, database, etc.
    result := fmt.Sprintf(`{"temperature": 22, "city": "%s", "condition": "sunny"}`, args.City)
    return result, nil
}

// 2. Create an executor
executor := tools.NewMap(map[string]tools.Func{
    "get_weather": weatherFunc,
})

// 3. Define tool schema (what the LLM sees)
toolDefs := []langrails.ToolDefinition{
    {
        Name:        "get_weather",
        Description: "Get the current weather for a city",
        Parameters: json.RawMessage(`{
            "type": "object",
            "properties": {
                "city": {"type": "string", "description": "City name"}
            },
            "required": ["city"]
        }`),
    },
}
```

## Running the Tool Loop

```go
req := &langrails.CompletionRequest{
    Model:        "gpt-4o",
    SystemPrompt: "You are a helpful assistant with access to tools.",
    Messages:     messages,
    Tools:        toolDefs,
}

result, err := tools.RunLoop(ctx, provider, req, executor)
if err != nil {
    log.Fatal(err)
}

fmt.Println(result.Response.Content)
fmt.Printf("Iterations: %d, Tokens: %d\n", result.Iterations, result.TotalUsage.TotalTokens)
```

The loop runs until the LLM returns a text response (no more tool calls) or hits the max iterations.

## Options

```go
result, err := tools.RunLoop(ctx, provider, req, executor,
    // Limit iterations (default: 20)
    tools.WithMaxIterations(5),

    // Hook for logging/tracing each tool call
    tools.WithToolCallHook(func(call langrails.ToolCall, result string, err error) {
        fmt.Printf("Tool: %s, Result: %s\n", call.Name, result)
    }),
)
```

## Multiple Tools

Register as many tools as you need. The LLM decides which to call (and can call multiple in one turn):

```go
executor := tools.NewMap(map[string]tools.Func{
    "get_weather":    weatherFunc,
    "search_web":     searchFunc,
    "calculate":      calcFunc,
    "send_email":     emailFunc,
    "query_database": dbFunc,
})
```

## Tool Response Format

Tools return a string — typically JSON. The LLM parses it and uses the data in its response:

```go
// Good: structured JSON
return `{"temperature": 22, "unit": "celsius", "condition": "sunny"}`, nil

// Good: simple text
return "The file was saved successfully.", nil

// Errors: return an error, the framework wraps it for the LLM
return "", fmt.Errorf("city not found: %s", args.City)
```

## JSON Schema Tips

The `Parameters` field uses [JSON Schema](https://json-schema.org/) to describe what arguments the tool accepts:

```go
// String parameter
{"type": "string", "description": "Search query"}

// Number with constraints
{"type": "number", "minimum": 0, "maximum": 100}

// Enum (fixed choices)
{"type": "string", "enum": ["celsius", "fahrenheit"]}

// Optional parameters: don't include in "required"
{
    "type": "object",
    "properties": {
        "query": {"type": "string"},
        "limit": {"type": "integer", "description": "Max results (default 10)"}
    },
    "required": ["query"]
}
```

## Example: Web Search Tool

```go
searchFunc := func(ctx context.Context, arguments string) (string, error) {
    var args struct {
        Query string `json:"query"`
        Limit int    `json:"limit"`
    }
    json.Unmarshal([]byte(arguments), &args)
    if args.Limit == 0 {
        args.Limit = 5
    }

    // Call your search API
    results, err := searchAPI.Search(ctx, args.Query, args.Limit)
    if err != nil {
        return "", err
    }

    out, _ := json.Marshal(results)
    return string(out), nil
}
```

## Integrating with the Demo App

To add custom tools to the demo chat app, edit `internal/engine/chat.go`:

1. Create your tool functions and definitions
2. Add them to the executor alongside media tools
3. Include the definitions in the completion request

The media tools in `internal/engine/media_tools.go` are a good reference implementation.
