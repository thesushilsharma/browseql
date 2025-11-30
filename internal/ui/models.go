package ui

import (
	"browseql/internal/database"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *model) InitProgram() tea.Model {
	return m
}

func (m model) Init() tea.Cmd {
	return m.loadTables()
}

// Messages
type tablesLoadedMsg struct {
	tables []string
}

type errorMsg struct {
	message string
}

func (m *model) loadTables() tea.Cmd {
	return func() tea.Msg {
		tables, err := m.dbManager.GetTables()
		if err != nil {
			return errorMsg{message: err.Error()}
		}
		return tablesLoadedMsg{tables: tables}
	}
}

// Update model struct
type model struct {
	dbManager     *database.DBManager
	tables        []string
	tableData     [][]string
	headers       []string
	cursor        int
	loading       bool
	errorMsg      string
	selectedTable string
	mode          string
	viewport      viewport.Model // Add this
	width         int
	height        int
}

// Add new message types
type dataLoadedMsg struct {
	tableName string
	headers   []string
	data      [][]string
}

// Add styles
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true)
)

// Update NewModel
func NewModel(dbManager *database.DBManager) *model {
	vp := viewport.New(80, 20)
	return &model{
		dbManager: dbManager,
		loading:   true,
		viewport:  vp,
		mode:      "tables",
	}
}

// Update Update method to handle window size
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// -------------------------
	// Handle window resizing
	// -------------------------
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 4
		return m, nil

	// -------------------------
	// Handle key events
	// -------------------------
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "esc":
			if m.mode == "data" {
				m.mode = "tables"
				return m, nil
			}
		}

		// Mode-specific key handling
		switch m.mode {
		case "tables":
			return m.updateTablesMode(msg)

		case "data":
			return m.updateDataMode(msg)
		}

	// -------------------------
	// Async messages
	// -------------------------
	case tablesLoadedMsg:
		m.loading = false
		m.tables = msg.tables

	case dataLoadedMsg:
		m.loading = false
		m.headers = msg.headers
		m.tableData = msg.data
		m.selectedTable = msg.tableName
		m.mode = "data"

	case errorMsg:
		m.loading = false
		m.errorMsg = msg.message
	}

	return m, nil
}

// Update data mode update
func (m model) updateDataMode(msg tea.KeyMsg) (model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.String() {
	case "up", "k":
		m.viewport.LineUp(1)
	case "down", "j":
		m.viewport.LineDown(1)
	}
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// Update View method
func (m model) View() string {
	if m.loading {
		return "Loading...\n\nPress q to quit"
	}

	if m.errorMsg != "" {
		return fmt.Sprintf("Error: %s\n\nPress q to quit", m.errorMsg)
	}

	var s strings.Builder

	switch m.mode {
	case "tables":
		s.WriteString("🗄️  Database Tables\n\n")
		for i, table := range m.tables {
			cursor := " "
			if i == m.cursor {
				cursor = ">"
			}
			s.WriteString(fmt.Sprintf("%s %s\n", cursor, table))
		}
		s.WriteString("\nEnter: Select table • q: Quit")

	case "data":
		s.WriteString(fmt.Sprintf("Table: %s\n\n", m.selectedTable))
		if len(m.tableData) > 0 {
			// Display headers
			for i, header := range m.headers {
				if i > 0 {
					s.WriteString(" | ")
				}
				s.WriteString(header)
			}
			s.WriteString("\n")
			s.WriteString(strings.Repeat("-", len(m.headers)*15))
			s.WriteString("\n")

			// Display data (first 5 rows for now)
			for i, row := range m.tableData {
				if i == 0 { // Skip header row
					continue
				}
				if i > 5 { // Limit to 5 rows for now
					break
				}
				for j, cell := range row {
					if j > 0 {
						s.WriteString(" | ")
					}
					s.WriteString(cell)
				}
				s.WriteString("\n")
			}
		}
		s.WriteString("\nEsc: Back • q: Quit")
	}

	return s.String()
}

func (m model) renderTables() string {
	var s strings.Builder
	s.WriteString(titleStyle.Render("🗄️  Database Tables"))
	s.WriteString("\n\n")

	for i, table := range m.tables {
		cursor := "  "
		if i == m.cursor {
			cursor = selectedStyle.Render("> ")
		}
		s.WriteString(fmt.Sprintf("%s%s\n", cursor, table))
	}

	s.WriteString("\n" + m.renderHelp())
	return s.String()
}

func (m model) renderData() string {
	var s strings.Builder
	s.WriteString(titleStyle.Render(fmt.Sprintf("Table: %s", m.selectedTable)))
	s.WriteString("\n\n")

	// Simple table display for now
	if len(m.tableData) > 0 {
		for i, row := range m.tableData {
			for j, cell := range row {
				if j > 0 {
					s.WriteString(" | ")
				}
				if i == 0 { // Header row
					s.WriteString(lipgloss.NewStyle().Bold(true).Render(cell))
				} else {
					s.WriteString(cell)
				}
			}
			s.WriteString("\n")
			if i == 0 {
				s.WriteString(strings.Repeat("-", 50))
				s.WriteString("\n")
			}
		}
	}

	s.WriteString("\n" + m.renderHelp())
	return s.String()
}

func (m model) renderHelp() string {
	switch m.mode {
	case "tables":
		return "↑↓: Navigate • Enter: Select • q: Quit"
	case "data":
		return "↑↓: Scroll • Esc: Back • q: Quit"
	default:
		return "q: Quit"
	}
}

// Add this method
func (m *model) loadTableData(tableName string) tea.Cmd {
	return func() tea.Msg {
		data, headers, err := m.dbManager.GetTableData(tableName, 100)
		if err != nil {
			return errorMsg{message: err.Error()}
		}
		return dataLoadedMsg{
			tableName: tableName,
			headers:   headers,
			data:      data,
		}
	}
}

func (m model) updateTablesMode(msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.tables)-1 {
			m.cursor++
		}
	case "enter", " ":
		if len(m.tables) > 0 {
			return m, m.loadTableData(m.tables[m.cursor])
		}
	}
	return m, nil
}
