package engine

import (
	"database/sql"
	"time"
)

// ChatMessage is a persisted chat message.
type ChatMessage struct {
	Role      string
	Content   string
	CreatedAt time.Time
}

// History manages persistent chat history in SQLite.
type History struct {
	db *sql.DB
}

// NewHistory creates a history store, initializing the table if needed.
func NewHistory(db *sql.DB) (*History, error) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS chat_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT (datetime('now'))
		)
	`)
	if err != nil {
		return nil, err
	}
	return &History{db: db}, nil
}

// Add stores a message.
func (h *History) Add(role, content string) error {
	_, err := h.db.Exec("INSERT INTO chat_history (role, content) VALUES (?, ?)", role, content)
	return err
}

// Load returns the last n messages.
func (h *History) Load(n int) ([]ChatMessage, error) {
	rows, err := h.db.Query(
		"SELECT role, content, created_at FROM chat_history ORDER BY id DESC LIMIT ?", n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []ChatMessage
	for rows.Next() {
		var m ChatMessage
		var ts string
		if err := rows.Scan(&m.Role, &m.Content, &ts); err != nil {
			continue
		}
		m.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", ts)
		msgs = append(msgs, m)
	}

	// Reverse to chronological order
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}

	return msgs, nil
}

// Clear deletes all history.
func (h *History) Clear() error {
	_, err := h.db.Exec("DELETE FROM chat_history")
	return err
}
