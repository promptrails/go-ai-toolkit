// Command chat is an AI chat TUI powered by Go AI Toolkit.
// It demonstrates LangRails, GuardRails, and MemoryRails working together.
package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/promptrails/go-ai-toolkit/internal/config"
	"github.com/promptrails/go-ai-toolkit/internal/engine"
	"github.com/promptrails/go-ai-toolkit/internal/tui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n\nUsage:\n  export OPENAI_API_KEY=sk-...\n  ai-chat\n", err)
		os.Exit(1)
	}

	// Logger writes to file in data dir (not stdout — TUI owns stdout)
	logger := initLogger(cfg.DataDir)
	defer func() { _ = logger.Sync() }()

	logger.Info("starting ai-chat",
		zap.String("model", cfg.Model),
		zap.Bool("memory", cfg.Memory),
		zap.String("data_dir", cfg.DataDir))

	db, err := sql.Open("sqlite3", cfg.DBPath())
	if err != nil {
		logger.Fatal("failed to open database", zap.Error(err))
	}
	defer db.Close()

	chat, err := engine.NewChat(cfg, db, logger)
	if err != nil {
		logger.Fatal("failed to init engine", zap.Error(err))
	}

	model := tui.New(chat)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		logger.Fatal("TUI error", zap.Error(err))
	}
}

// initLogger creates a zap logger that writes JSON to a log file.
func initLogger(dataDir string) *zap.Logger {
	logPath := filepath.Join(dataDir, "ai-chat.log")

	cfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(zapcore.DebugLevel),
		Development: false,
		Encoding:    "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:      "ts",
			LevelKey:     "level",
			MessageKey:   "msg",
			CallerKey:    "caller",
			EncodeTime:   zapcore.ISO8601TimeEncoder,
			EncodeLevel:  zapcore.LowercaseLevelEncoder,
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
		OutputPaths: []string{logPath},
	}

	logger, err := cfg.Build()
	if err != nil {
		// Fallback to nop logger if file logging fails
		return zap.NewNop()
	}
	return logger
}
