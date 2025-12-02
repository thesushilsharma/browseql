package database

import (
	"database/sql"
	"fmt"
	"strings"

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

func (m *DBManager) GetTableData(tableName string, limit int) ([][]string, []string, error) {
	// Get column names
	query := fmt.Sprintf("SELECT * FROM %s LIMIT 1", tableName)
	rows, err := m.db.Query(query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}

	// Get data
	query = fmt.Sprintf("SELECT * FROM %s LIMIT ?", tableName)
	dataRows, err := m.db.Query(query, limit)
	if err != nil {
		return nil, nil, err
	}
	defer dataRows.Close()

	var data [][]string
	data = append(data, columns) // Headers as first row

	for dataRows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := dataRows.Scan(valuePtrs...); err != nil {
			return nil, nil, err
		}

		row := make([]string, len(columns))
		for i, val := range values {
			if val == nil {
				row[i] = "NULL"
			} else {
				row[i] = fmt.Sprintf("%v", val)
			}
		}
		data = append(data, row)
	}

	return data, columns, nil
}

func (m *DBManager) ExecuteQuery(query string) ([]string, [][]string, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil, fmt.Errorf("empty query")
	}

	// Check if it's a SELECT query
	upperQuery := strings.ToUpper(query)
	if strings.HasPrefix(upperQuery, "SELECT") {
		rows, err := m.db.Query(query)
		if err != nil {
			return nil, nil, err
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			return nil, nil, err
		}

		var data [][]string
		data = append(data, columns)

		for rows.Next() {
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range columns {
				valuePtrs[i] = &values[i]
			}

			if err := rows.Scan(valuePtrs...); err != nil {
				return nil, nil, err
			}

			row := make([]string, len(columns))
			for i, val := range values {
				if val == nil {
					row[i] = "NULL"
				} else {
					row[i] = fmt.Sprintf("%v", val)
				}
			}
			data = append(data, row)
		}

		return columns, data, nil
	} else {
		result, err := m.db.Exec(query)
		if err != nil {
			return nil, nil, err
		}

		rowsAffected, _ := result.RowsAffected()
		columns := []string{"Result"}
		data := [][]string{
			{"Query executed successfully"},
			{fmt.Sprintf("Rows affected: %d", rowsAffected)},
		}

		return columns, data, nil
	}
}