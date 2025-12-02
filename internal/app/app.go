package app

import (
	"fmt"
	// "os"

	tea "github.com/charmbracelet/bubbletea"
	"browseql/internal/database"
	"browseql/internal/ui"
)

type App struct {
	dbPath string
}

func NewApp(dbPath string) *App {
	return &App{dbPath: dbPath}
}

func (a *App) Run() error {
	// Initialize database connection
	dbManager, err := database.NewManager(a.dbPath)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer dbManager.Close()

	// Initialize the UI model
	model := ui.NewModel(dbManager)

	// Start the Bubble Tea program
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),       // Use alternate screen buffer
		tea.WithMouseCellMotion(), // Enable mouse support
	)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("program error: %w", err)
	}

	return nil
}