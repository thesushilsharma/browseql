package ui

import (
	"github.com/charmbracelet/bubbles/textinput"
)

// newQueryModel creates and configures a text input for SQL queries
func newQueryModel() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "Enter SQL query (e.g., SELECT * FROM users)..."
	ti.Prompt = "SQL> "
	ti.Focus()
	ti.CharLimit = 1000
	ti.Width = 50
	return ti
}