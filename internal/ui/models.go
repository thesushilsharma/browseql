package ui

import (
	"fmt"
	"strings"

	"browseql/internal/database"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ====================
// Message Types
// ====================
type tablesLoadedMsg struct {
	tables []string
}

type dataLoadedMsg struct {
	tableName string
	headers   []string
	data      [][]string
}

type errorMsg struct {
	message string
}

// ====================
// Model Definition
// ====================
type model struct {
	dbManager     *database.DBManager
	tables        []string
	tableData     [][]string
	headers       []string
	cursor        int
	loading       bool
	errorMsg      string
	selectedTable string
	mode          string // "tables", "data", or "query"
	viewport      viewport.Model
	queryInput    textinput.Model
	width         int
	height        int
}

// ====================
// Styles
// ====================
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B9D")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))
)

// ====================
// Model Methods
// ====================

// NewModel creates a new UI model
func NewModel(dbManager *database.DBManager) tea.Model {
	vp := viewport.New(80, 20)
	qi := newQueryModel()
	
	return &model{
		dbManager:  dbManager,
		loading:    true,
		viewport:   vp,
		queryInput: qi,
		mode:       "tables",
	}
}

// Init initializes the model
func (m model) Init() tea.Cmd {
	return m.loadTables()
}

// Update handles messages and updates the model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 6
		if m.mode == "query" {
			m.queryInput.Width = msg.Width - 10
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			if m.mode == "data" || m.mode == "query" {
				m.mode = "tables"
				m.queryInput.SetValue("")
				return m, nil
			}
		case ":":
			if m.mode == "tables" {
				m.mode = "query"
				m.queryInput.Focus()
				return m, nil
			}
		case "r":
			if m.mode == "data" && m.selectedTable != "" {
				m.loading = true
				return m, m.loadTableData(m.selectedTable)
			}
		}

		// Mode-specific key handling
		switch m.mode {
		case "tables":
			return m.updateTablesMode(msg)
		case "data":
			return m.updateDataMode(msg)
		case "query":
			return m.updateQueryMode(msg)
		}

	case tablesLoadedMsg:
		m.loading = false
		m.tables = msg.tables
		return m, nil

	case dataLoadedMsg:
		m.loading = false
		m.headers = msg.headers
		m.tableData = msg.data
		m.selectedTable = msg.tableName
		m.mode = "data"
		m.updateViewport()
		return m, nil

	case errorMsg:
		m.loading = false
		m.errorMsg = msg.message
		return m, nil
	}

	return m, nil
}

// View renders the UI
func (m model) View() string {
	if m.loading {
		return m.renderLoading()
	}

	if m.errorMsg != "" {
		return m.renderError()
	}

	switch m.mode {
	case "tables":
		return m.renderTables()
	case "data":
		return m.renderData()
	case "query":
		return m.renderQuery()
	default:
		return "Unknown mode"
	}
}

// ====================
// Helper Methods
// ====================

func (m *model) loadTables() tea.Cmd {
	return func() tea.Msg {
		tables, err := m.dbManager.GetTables()
		if err != nil {
			return errorMsg{message: err.Error()}
		}
		return tablesLoadedMsg{tables: tables}
	}
}

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

func (m model) updateTablesMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
			m.loading = true
			return m, m.loadTableData(m.tables[m.cursor])
		}
	}
	return m, nil
}

func (m model) updateDataMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		m.viewport.LineUp(1)
	case "down", "j":
		m.viewport.LineDown(1)
	}
	// Remove unused cmd variable
	m.viewport, _ = m.viewport.Update(msg)
	return m, nil
}

func (m model) updateQueryMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		query := m.queryInput.Value()
		if strings.TrimSpace(query) != "" {
			m.queryInput.SetValue("")
			m.loading = true
			return m, m.executeQuery(query)
		}
	case "esc":
		m.mode = "tables"
		m.queryInput.SetValue("")
		return m, nil
	}

	// Update the text input
	var cmd tea.Cmd
	m.queryInput, cmd = m.queryInput.Update(msg)
	return m, cmd
}

func (m *model) executeQuery(query string) tea.Cmd {
	return func() tea.Msg {
		query = strings.TrimSpace(query)
		if query == "" {
			return errorMsg{message: "Query cannot be empty"}
		}

		headers, data, err := m.dbManager.ExecuteQuery(query)
		if err != nil {
			return errorMsg{message: "SQL Error: " + err.Error()}
		}

		return dataLoadedMsg{
			tableName: "Query Results",
			headers:   headers,
			data:      data,
		}
	}
}

// ====================
// Rendering Methods
// ====================

func (m model) renderLoading() string {
	var s strings.Builder
	s.WriteString(titleStyle.Render("🗄️ BrowseQL"))
	s.WriteString("\n\n")
	s.WriteString("Loading...")
	s.WriteString("\n\n")
	s.WriteString(helpStyle.Render("q: Quit"))
	return s.String()
}

func (m model) renderError() string {
	var s strings.Builder
	s.WriteString(errorStyle.Render("Error: " + m.errorMsg))
	s.WriteString("\n\n")
	s.WriteString(m.renderHelp())
	return s.String()
}

func (m model) renderTables() string {
	var s strings.Builder
	s.WriteString(titleStyle.Render("🗄️ Database Tables"))
	s.WriteString("\n\n")

	if len(m.tables) == 0 {
		s.WriteString("No tables found in database\n")
	} else {
		for i, table := range m.tables {
			cursor := "  "
			if i == m.cursor {
				cursor = selectedStyle.Render("> ")
			}
			// Simplified string concatenation
			s.WriteString(cursor + table + "\n")
		}
	}

	s.WriteString("\n" + m.renderHelp())
	return s.String()
}

func (m model) renderData() string {
	var s strings.Builder
	// Simplified fmt.Sprintf usage
	title := fmt.Sprintf("📊 Table: %s (%d rows)", m.selectedTable, len(m.tableData)-1)
	s.WriteString(titleStyle.Render(title))
	s.WriteString("\n\n")
	
	s.WriteString(m.viewport.View())
	
	s.WriteString("\n" + m.renderHelp())
	return s.String()
}

func (m model) renderQuery() string {
	var s strings.Builder
	s.WriteString(titleStyle.Render("🔍 SQL Query Mode"))
	s.WriteString("\n\n")
	
	s.WriteString("Enter SQL query and press Enter:\n\n")
	s.WriteString(m.queryInput.View())
	
	currentQuery := m.queryInput.Value()
	if currentQuery != "" {
		s.WriteString("\n\nQuery: ")
		s.WriteString(lipgloss.NewStyle().Italic(true).Render(currentQuery))
	}
	
	s.WriteString("\n\n" + m.renderHelp())
	return s.String()
}

func (m model) renderHelp() string {
	switch m.mode {
	case "tables":
		return helpStyle.Render("↑↓: Navigate • Enter: Select • :: Query • q: Quit")
	case "data":
		return helpStyle.Render("↑↓: Scroll • r: Refresh • Esc: Back • :: Query • q: Quit")
	case "query":
		return helpStyle.Render("Enter: Execute • Esc: Cancel • q: Quit")
	default:
		return helpStyle.Render("q: Quit")
	}
}

func (m *model) updateViewport() {
	if len(m.tableData) == 0 {
		m.viewport.SetContent("No data available")
		return
	}
	
	var content strings.Builder
	
	// Calculate column widths
	colWidths := make([]int, len(m.headers))
	for i, header := range m.headers {
		colWidths[i] = len(header)
	}
	
	for _, row := range m.tableData {
		for i, cell := range row {
			if len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}
	
	// Limit column widths
	for i := range colWidths {
		if colWidths[i] > 30 {
			colWidths[i] = 30
		}
	}
	
	// Build header
	headerRow := "│"
	for i, header := range m.headers {
		padded := padString(header, colWidths[i])
		headerRow += " " + lipgloss.NewStyle().Bold(true).Render(padded) + " │"
	}
	content.WriteString(headerRow + "\n")
	
	// Separator
	separator := "├"
	for i, width := range colWidths {
		separator += strings.Repeat("─", width+2)
		if i < len(colWidths)-1 {
			separator += "┼"
		}
	}
	separator += "┤\n"
	content.WriteString(separator)
	
	// Build rows
	for i, row := range m.tableData {
		if i == 0 { // Skip header row
			continue
		}
		rowStr := "│"
		for j, cell := range row {
			padded := padString(cell, colWidths[j])
			rowStr += " " + padded + " │"
		}
		content.WriteString(rowStr + "\n")
	}
	
	m.viewport.SetContent(content.String())
}

// Utility function
func padString(s string, width int) string {
	if len(s) > width {
		return s[:width-3] + "..."
	}
	return s + strings.Repeat(" ", width-len(s))
}