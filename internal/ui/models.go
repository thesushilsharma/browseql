package ui

import (
	"browseql/internal/database"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

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
    mode          string // "tables" or "data"
}

// Add new message types
type dataLoadedMsg struct {
    tableName string
    headers   []string
    data      [][]string
}

// Update Update method
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
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

        switch m.mode {
        case "tables":
            return m.updateTablesMode(msg)
        case "data":
            return m.updateDataMode(msg)
        }
    
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

func (m model) updateDataMode(msg tea.KeyMsg) (model, tea.Cmd) {
    // For now, just handle escape (already handled above)
    return m, nil
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