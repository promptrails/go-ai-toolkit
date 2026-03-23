package engine

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/promptrails/langrails"
	"github.com/promptrails/langrails/tools"
	"github.com/promptrails/mediarails"
	"github.com/promptrails/mediarails/media"
	"go.uber.org/zap"

	"github.com/promptrails/go-ai-toolkit/internal/config"
)

// Default model IDs for media generation.
const (
	dalleModel      = "dall-e-3"
	elevenLabsModel = "eleven_multilingual_v2"
	falVideoModel   = "fal-ai/minimax-video"

	defaultElevenLabsVoice = "21m00Tcm4TlvDq8ikWAM" // Rachel
)

// mediaToolKit holds media providers and tool definitions, built from config.
type mediaToolKit struct {
	executor tools.Executor
	defs     []langrails.ToolDefinition
}

// toolParams builds a JSON Schema parameters object from a map of properties and required fields.
func toolParams(properties map[string]any, required []string) json.RawMessage {
	schema := map[string]any{
		"type":       "object",
		"properties": properties,
		"required":   required,
	}
	b, _ := json.Marshal(schema)
	return b
}

// buildMediaTools creates media tool definitions and executor based on available API keys.
func buildMediaTools(cfg *config.Config, logger *zap.Logger) *mediaToolKit {
	funcs := map[string]tools.Func{}
	var defs []langrails.ToolDefinition

	// generate_image — always available (OpenAI DALL-E via OPENAI_API_KEY)
	{
		p, err := media.New(media.OpenAIImage, cfg.APIKey)
		if err == nil {
			funcs["generate_image"] = makeImageTool(p)
			defs = append(defs, langrails.ToolDefinition{
				Name:        "generate_image",
				Description: "Generate an image from a text description using DALL-E. Returns the image URL.",
				Parameters: toolParams(map[string]any{
					"prompt":  map[string]any{"type": "string", "description": "Description of the image to generate"},
					"quality": map[string]any{"type": "string", "enum": []string{"standard", "hd"}, "description": "Image quality (default: standard)"},
				}, []string{"prompt"}),
			})
			logger.Info("media tool registered", zap.String("tool", "generate_image"), zap.String("provider", "openai"))
		}
	}

	// text_to_speech — only if ELEVENLABS_API_KEY is set
	if cfg.ElevenLabsAPIKey != "" {
		p, err := media.New(media.ElevenLabs, cfg.ElevenLabsAPIKey)
		if err == nil {
			funcs["text_to_speech"] = makeTTSTool(p)
			defs = append(defs, langrails.ToolDefinition{
				Name:        "text_to_speech",
				Description: "Convert text to speech audio using ElevenLabs. Returns the audio URL.",
				Parameters: toolParams(map[string]any{
					"text": map[string]any{"type": "string", "description": "Text to convert to speech"},
				}, []string{"text"}),
			})
			logger.Info("media tool registered", zap.String("tool", "text_to_speech"), zap.String("provider", "elevenlabs"))
		}
	}

	// generate_video — only if FAL_API_KEY is set
	if cfg.FalAPIKey != "" {
		p, err := media.New(media.Fal, cfg.FalAPIKey)
		if err == nil {
			funcs["generate_video"] = makeVideoTool(p)
			defs = append(defs, langrails.ToolDefinition{
				Name:        "generate_video",
				Description: "Generate a short video from a text description using Fal AI. Returns the video URL.",
				Parameters: toolParams(map[string]any{
					"prompt": map[string]any{"type": "string", "description": "Description of the video to generate"},
				}, []string{"prompt"}),
			})
			logger.Info("media tool registered", zap.String("tool", "generate_video"), zap.String("provider", "fal"))
		}
	}

	if len(funcs) == 0 {
		return nil
	}

	return &mediaToolKit{
		executor: tools.NewMap(funcs),
		defs:     defs,
	}
}

type imageArgs struct {
	Prompt  string `json:"prompt"`
	Quality string `json:"quality,omitempty"`
}

func makeImageTool(p mediarails.Provider) tools.Func {
	return func(ctx context.Context, arguments string) (string, error) {
		var args imageArgs
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return "", fmt.Errorf("invalid arguments: %w", err)
		}

		cfg := map[string]any{}
		if args.Quality != "" {
			cfg["quality"] = args.Quality
		}

		resp, err := p.Generate(ctx, &mediarails.GenerateRequest{
			Type:   mediarails.ImageGen,
			Model:  dalleModel,
			Prompt: args.Prompt,
			Config: cfg,
		})
		if err != nil {
			return "", fmt.Errorf("image generation failed: %w", err)
		}

		result := map[string]string{"url": resp.AssetURL, "status": "completed"}
		out, _ := json.Marshal(result)
		return string(out), nil
	}
}

type ttsArgs struct {
	Text string `json:"text"`
}

func makeTTSTool(p mediarails.Provider) tools.Func {
	return func(ctx context.Context, arguments string) (string, error) {
		var args ttsArgs
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return "", fmt.Errorf("invalid arguments: %w", err)
		}

		resp, err := p.Generate(ctx, &mediarails.GenerateRequest{
			Type:   mediarails.TTS,
			Model:  elevenLabsModel,
			Prompt: args.Text,
			Config: map[string]any{"voice_id": defaultElevenLabsVoice},
		})
		if err != nil {
			return "", fmt.Errorf("TTS failed: %w", err)
		}

		// ElevenLabs returns bytes, not URL
		if resp.AssetURL != "" {
			result := map[string]string{"url": resp.AssetURL, "status": "completed"}
			out, _ := json.Marshal(result)
			return string(out), nil
		}

		return `{"status": "completed", "message": "Audio generated (binary output)"}`, nil
	}
}

type videoArgs struct {
	Prompt string `json:"prompt"`
}

func makeVideoTool(p mediarails.Provider) tools.Func {
	return func(ctx context.Context, arguments string) (string, error) {
		var args videoArgs
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return "", fmt.Errorf("invalid arguments: %w", err)
		}

		resp, err := p.Generate(ctx, &mediarails.GenerateRequest{
			Type:   mediarails.VideoGen,
			Model:  falVideoModel,
			Prompt: args.Prompt,
		})
		if err != nil {
			return "", fmt.Errorf("video generation failed: %w", err)
		}

		result := map[string]string{
			"status": string(resp.Status),
		}
		if resp.AssetURL != "" {
			result["url"] = resp.AssetURL
		}
		if resp.JobID != "" {
			result["job_id"] = resp.JobID
			result["message"] = "Video generation started. It may take a few minutes to complete."
		}

		out, _ := json.Marshal(result)
		return string(out), nil
	}
}
