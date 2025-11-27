package ui

import (
	"browseql/internal/database"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
    dbManager *database.DBManager
    tables    []string
    cursor    int
    loading   bool
    errorMsg  string
}

func NewModel(dbManager *database.DBManager) *model {
    return &model{
        dbManager: dbManager,
        loading:   true,
    }
}

func (m *model) InitProgram() tea.Model {
    return m
}

func (m model) Init() tea.Cmd {
    return m.loadTables()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q":
            return m, tea.Quit
        case "up", "k":
            if m.cursor > 0 {
                m.cursor--
            }
        case "down", "j":
            if m.cursor < len(m.tables)-1 {
                m.cursor++
            }
        }
    case tablesLoadedMsg:
        m.loading = false
        m.tables = msg.tables
    case errorMsg:
        m.loading = false
        m.errorMsg = msg.message
    }
    return m, nil
}

func (m model) View() string {
    if m.loading {
        return "Loading tables...\n\nPress q to quit"
    }

    if m.errorMsg != "" {
        return fmt.Sprintf("Error: %s\n\nPress q to quit", m.errorMsg)
    }

    var s strings.Builder
    s.WriteString("🗄️  Database Tables\n\n")

    for i, table := range m.tables {
        cursor := " "
        if i == m.cursor {
            cursor = ">"
        }
        s.WriteString(fmt.Sprintf("%s %s\n", cursor, table))
    }

    s.WriteString("\nPress q to quit")
    return s.String()
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