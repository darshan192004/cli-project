package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteDB struct {
	DB     *sql.DB
	dbPath string
}

func ConnectSQLite(dbPath string) (*SQLiteDB, error) {
	if dbPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		dsDir := filepath.Join(home, ".dataset-cli")
		if err := os.MkdirAll(dsDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create data directory: %w", err)
		}
		dbPath = filepath.Join(dsDir, "dataset.db")
	}

	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_foreign_keys=ON")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db.SetMaxOpenConns(1)

	return &SQLiteDB{
		DB:     db,
		dbPath: dbPath,
	}, nil
}

func (s *SQLiteDB) Close() error {
	if s.DB != nil {
		return s.DB.Close()
	}
	return nil
}

func (s *SQLiteDB) Type() BackendType { return BackendSQLite }

func (s *SQLiteDB) GetDBPath() string {
	return s.dbPath
}

func (s *SQLiteDB) TableExists(ctx context.Context, tableName string) (bool, error) {
	query := `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`
	var count int
	err := s.DB.QueryRowContext(ctx, query, tableName).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *SQLiteDB) GetTableSchema(ctx context.Context, tableName string) ([]ColumnInfo, error) {
	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	rows, err := s.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		var cid int
		var notnull, pk int
		var defaultVal sql.NullString
		if err := rows.Scan(&cid, &col.Name, &col.DataType, &notnull, &defaultVal, &pk); err != nil {
			return nil, err
		}
		if defaultVal.Valid {
			col.Default = &defaultVal.String
		}
		if notnull == 1 {
			col.IsNullable = "NO"
		} else {
			col.IsNullable = "YES"
		}
		columns = append(columns, col)
	}
	return columns, nil
}

func (s *SQLiteDB) GetTableInfo(ctx context.Context, tableName string) (*TableInfo, error) {
	exists, err := s.TableExists(ctx, tableName)
	if err != nil || !exists {
		return nil, fmt.Errorf("table '%s' does not exist", tableName)
	}

	columns, err := s.GetTableSchema(ctx, tableName)
	if err != nil {
		return nil, err
	}

	var count int64
	err = s.DB.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&count)
	if err != nil {
		return nil, err
	}

	return &TableInfo{
		Name:    tableName,
		Columns: columns,
		Count:   count,
	}, nil
}

func (s *SQLiteDB) GetAllTables(ctx context.Context) ([]string, error) {
	query := `SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`
	rows, err := s.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

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

func (s *SQLiteDB) Execute(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	return results, nil
}

func (s *SQLiteDB) Exec(ctx context.Context, query string, args ...interface{}) (int64, error) {
	result, err := s.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (s *SQLiteDB) QueryRow(ctx context.Context, query string, args ...interface{}) (interface{}, error) {
	return s.DB.QueryRowContext(ctx, query, args...), nil
}
