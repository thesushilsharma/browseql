package app

import (
	"browseql/internal/database"
	"browseql/internal/ui"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type App struct {
	dbPath string
}

func NewApp(dbPath string) *App {
	return &App{dbPath: dbPath}
}

func (a *App) Run() error {
	dbManager, err := database.NewManager(a.dbPath)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer dbManager.Close()

	model := ui.NewModel(dbManager)
	p := tea.NewProgram(model)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("program error: %w", err)
	}
	return nil
}
