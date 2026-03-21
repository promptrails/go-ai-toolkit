package engine

import "context"

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
}
