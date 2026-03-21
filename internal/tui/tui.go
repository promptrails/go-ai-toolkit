package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/promptrails/go-ai-toolkit/internal/engine"
)

var (
	userStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	assistantStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	systemStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Italic(true)
	errorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
	dimStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	titleStyle     = lipgloss.NewStyle().
			Background(lipgloss.Color("4")).
			Foreground(lipgloss.Color("15")).
			Bold(true).
			Padding(0, 1)
)

type responseMsg struct {
	content string
	err     error
}

// Model is the bubbletea model for the TUI.
type Model struct {
	engine   engine.Engine
	viewport viewport.Model
	textarea textarea.Model
	messages []string
	width    int
	height   int
	waiting  bool
}

// New creates a new TUI model.
func New(eng engine.Engine) Model {
	ta := textarea.New()
	ta.Placeholder = "Type a message... (/help for commands)"
	ta.Focus()
	ta.CharLimit = 4096
	ta.SetHeight(3)
	ta.ShowLineNumbers = false

	vp := viewport.New(80, 20)

	return Model{
		engine:   eng,
		viewport: vp,
		textarea: ta,
	}
}

func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyEnter:
			if m.waiting {
				break
			}

			input := strings.TrimSpace(m.textarea.Value())
			if input == "" {
				break
			}

			m.textarea.Reset()

			// Handle commands
			if m.engine.IsCommand(input) {
				cmd := m.engine.ParseCommand(input)
				result := m.engine.Execute(context.Background(), cmd)
				if result.Quit {
					return m, tea.Quit
				}
				m.addMessage(systemStyle.Render("system"), result.Response)
				m.viewport.GotoBottom()
				break
			}

			// Show user message
			m.addMessage(userStyle.Render("you"), input)
			m.viewport.GotoBottom()

			// Send to engine
			m.waiting = true
			return m, m.sendMessage(input)
		}

	case responseMsg:
		m.waiting = false
		if msg.err != nil {
			m.addMessage(errorStyle.Render("error"), msg.err.Error())
		} else {
			m.addMessage(assistantStyle.Render("ai"), msg.content)
		}
		m.viewport.GotoBottom()

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 7
		m.textarea.SetWidth(msg.Width - 2)
		m.updateViewportContent()
	}

	var vpCmd, taCmd tea.Cmd
	m.viewport, vpCmd = m.viewport.Update(msg)
	if !m.waiting {
		m.textarea, taCmd = m.textarea.Update(msg)
	}
	cmds = append(cmds, vpCmd, taCmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	title := titleStyle.Render(" AI Chat ") +
		dimStyle.Render(fmt.Sprintf(" model:%s", m.engine.Model()))

	if m.waiting {
		title += dimStyle.Render(" (thinking...)")
	}

	separator := dimStyle.Render(strings.Repeat("─", max(m.width, 40)))

	return fmt.Sprintf("%s\n%s\n%s\n%s",
		title,
		m.viewport.View(),
		separator,
		m.textarea.View(),
	)
}

func (m *Model) addMessage(role, content string) {
	wrapped := wordWrap(content, max(m.width-4, 40))
	m.messages = append(m.messages, fmt.Sprintf("%s: %s", role, wrapped))
	m.updateViewportContent()
}

func (m *Model) updateViewportContent() {
	content := strings.Join(m.messages, "\n\n")
	m.viewport.SetContent(content)
}

func (m *Model) sendMessage(input string) tea.Cmd {
	return func() tea.Msg {
		resp, err := m.engine.Send(context.Background(), input)
		return responseMsg{content: resp, err: err}
	}
}

func wordWrap(s string, width int) string {
	if width <= 0 || len(s) <= width {
		return s
	}
	var result strings.Builder
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if i > 0 {
			result.WriteString("\n")
		}
		words := strings.Fields(line)
		lineLen := 0
		for j, word := range words {
			if j > 0 && lineLen+1+len(word) > width {
				result.WriteString("\n")
				lineLen = 0
			} else if j > 0 {
				result.WriteString(" ")
				lineLen++
			}
			result.WriteString(word)
			lineLen += len(word)
		}
	}
	return result.String()
}
