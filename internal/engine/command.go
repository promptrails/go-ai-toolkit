package engine

// Command represents a slash command.
type Command string

const (
	CmdClear  Command = "/clear"
	CmdMemory Command = "/memory"
	CmdForget Command = "/forget"
	CmdHelp   Command = "/help"
	CmdQuit   Command = "/quit"
	CmdExit   Command = "/exit"
)

// CommandResult is the outcome of a command execution.
type CommandResult struct {
	// Response is the text to display.
	Response string

	// Quit signals the app should exit.
	Quit bool
}

// AllCommands returns all available commands with descriptions.
func AllCommands() map[Command]string {
	return map[Command]string{
		CmdClear:  "Clear chat history",
		CmdMemory: "Show stored memories",
		CmdForget: "Forget all memories",
		CmdHelp:   "Show this help",
		CmdQuit:   "Exit",
	}
}
