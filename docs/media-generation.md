# Media Generation

The toolkit includes AI media generation via LLM tool calling. When you ask the assistant to create an image, generate speech, or make a video, the LLM decides to call the appropriate tool and returns the result URL.

## How It Works

```
User: "Create an image of a sunset over mountains"
  │
  ├─→ LangRails: sends message to LLM with tool definitions
  │
  ├─→ LLM decides: call generate_image(prompt: "a sunset over mountains")
  │
  ├─→ MediaRails: DALL-E 3 generates the image
  │
  ├─→ LLM receives the URL, crafts a response
  │
  └─→ User sees: "Here's your image: https://..."
```

The LLM handles the decision-making — you don't need to write routing logic.

## Available Tools

| Tool | Provider | API Key | Always On |
|------|----------|---------|-----------|
| `generate_image` | OpenAI DALL-E 3 | `OPENAI_API_KEY` | Yes |
| `text_to_speech` | ElevenLabs | `ELEVENLABS_API_KEY` | No |
| `generate_video` | Fal AI | `FAL_API_KEY` | No |

Tools are registered automatically when their API key is set. No API key = no tool.

## Configuration

```bash
# Always available (uses your existing OpenAI key)
export OPENAI_API_KEY=sk-...

# Optional: enable TTS
export ELEVENLABS_API_KEY=your-key

# Optional: enable video generation
export FAL_API_KEY=your-key
```

## Image Generation (DALL-E 3)

Always available since it uses your existing OpenAI API key.

```
You: Generate an image of a cat wearing sunglasses on a beach
Assistant: Here's your image: https://oaidalleapiprodscus.blob.core.windows.net/...
```

The tool supports a `quality` parameter (`standard` or `hd`). The LLM chooses based on your request.

## Text-to-Speech (ElevenLabs)

Requires `ELEVENLABS_API_KEY`. Uses the `eleven_multilingual_v2` model with the Rachel voice.

```
You: Read this quote aloud: "The only way to do great work is to love what you do."
Assistant: Here's the audio: https://...
```

## Video Generation (Fal AI)

Requires `FAL_API_KEY`. Uses `fal-ai/minimax-video`. Video generation is async — the tool returns immediately with a job ID. The video URL is available once processing completes.

```
You: Create a short video of ocean waves at sunset
Assistant: Video generation started. It may take a few minutes to complete. Job ID: ...
```

## Under the Hood

Media tools use three packages together:

```go
// MediaRails creates the provider
provider, _ := media.New(media.OpenAIImage, apiKey)

// The tool function calls MediaRails
resp, _ := provider.Generate(ctx, &mediarails.GenerateRequest{
    Type:   mediarails.ImageGen,
    Model:  "dall-e-3",
    Prompt: "a cat in space",
})
// resp.AssetURL = "https://..."

// LangRails tools.RunLoop orchestrates the LLM ↔ tool loop
result, _ := tools.RunLoop(ctx, llmProvider, req, executor)
```

## Adding Your Own Media Tools

You can extend the toolkit with custom media tools by adding entries to the `mediaToolKit` in `internal/engine/media_tools.go`:

```go
// Register a new tool
funcs["my_tool"] = func(ctx context.Context, args string) (string, error) {
    // Parse args, call MediaRails, return URL
    return `{"url": "https://..."}`, nil
}

defs = append(defs, langrails.ToolDefinition{
    Name:        "my_tool",
    Description: "Description for the LLM",
    Parameters:  toolParams(map[string]any{
        "prompt": map[string]any{"type": "string", "description": "..."},
    }, []string{"prompt"}),
})
```
