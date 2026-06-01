package engine

import "context"

// Usage is the cumulative token consumption and estimated cost of a session.
// Cost is in USD and is only populated when model pricing is available from
// the models.dev catalog (HasCost reports whether it is).
type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	Cost             float64
	HasCost          bool
}

// Engine defines the chat business logic, independent of UI.
type Engine interface {
	// Send processes a user message and returns the assistant response.
	Send(ctx context.Context, input string) (string, error)

	// Execute runs a slash command and returns the result.
	Execute(ctx context.Context, cmd Command) CommandResult

	// IsCommand returns true if the input starts with /.
	IsCommand(input string) bool

	// ParseCommand extracts the command from input.
	ParseCommand(input string) Command

	// Model returns the current model name.
	Model() string

	// Usage returns the cumulative token usage and estimated cost so far.
	Usage() Usage
}
