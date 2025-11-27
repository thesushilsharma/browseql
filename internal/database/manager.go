package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type DBManager struct {
    db *sql.DB
}

func NewManager(dbPath string) (*DBManager, error) {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, err
    }

    if err := db.Ping(); err != nil {
        return nil, err
    }

    return &DBManager{db: db}, nil
}

func (m *DBManager) Close() error {
    if m.db != nil {
        return m.db.Close()
    }
    return nil
}

func (m *DBManager) GetTables() ([]string, error) {
    rows, err := m.db.Query(`
        SELECT name 
        FROM sqlite_master 
        WHERE type='table' AND name NOT LIKE 'sqlite_%'
        ORDER BY name
    `)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var tables []string
    for rows.Next() {
        var name string
        if err := rows.Scan(&name); err != nil {
            return nil, err
        }
        tables = append(tables, name)
    }
    return tables, nil
}