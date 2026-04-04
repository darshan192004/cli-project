package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/tursodatabase/libsql-client-go/libsql"
)

type LibSQLDB struct {
	DB  *sql.DB
	URL string
}

func ConnectLibSQL(url, authToken string) (*LibSQLDB, error) {
	if url == "" {
		return nil, fmt.Errorf("LIBSQL_URL environment variable is required for cloud sync")
	}

	opts := []libsql.Option{libsql.WithTls(true)}
	if authToken != "" {
		opts = append(opts, libsql.WithAuthToken(authToken))
	}

	connector, err := libsql.NewConnector(url, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create libsql connector: %w", err)
	}

	db := sql.OpenDB(connector)
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &LibSQLDB{
		DB:  db,
		URL: url,
	}, nil
}

func (l *LibSQLDB) Close() error {
	if l.DB != nil {
		return l.DB.Close()
	}
	return nil
}

func (l *LibSQLDB) Type() BackendType { return BackendLibSQL }

func (l *LibSQLDB) TableExists(ctx context.Context, tableName string) (bool, error) {
	query := `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`
	var count int
	err := l.DB.QueryRowContext(ctx, query, tableName).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (l *LibSQLDB) GetTableSchema(ctx context.Context, tableName string) ([]ColumnInfo, error) {
	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	rows, err := l.DB.QueryContext(ctx, query)
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

func (l *LibSQLDB) GetTableInfo(ctx context.Context, tableName string) (*TableInfo, error) {
	exists, err := l.TableExists(ctx, tableName)
	if err != nil || !exists {
		return nil, fmt.Errorf("table '%s' does not exist", tableName)
	}

	columns, err := l.GetTableSchema(ctx, tableName)
	if err != nil {
		return nil, err
	}

	var count int64
	err = l.DB.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&count)
	if err != nil {
		return nil, err
	}

	return &TableInfo{
		Name:    tableName,
		Columns: columns,
		Count:   count,
	}, nil
}

func (l *LibSQLDB) GetAllTables(ctx context.Context) ([]string, error) {
	query := `SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`
	rows, err := l.DB.QueryContext(ctx, query)
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

func (l *LibSQLDB) Execute(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := l.DB.QueryContext(ctx, query, args...)
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

func (l *LibSQLDB) Exec(ctx context.Context, query string, args ...interface{}) (int64, error) {
	result, err := l.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (l *LibSQLDB) QueryRow(ctx context.Context, query string, args ...interface{}) (interface{}, error) {
	return l.DB.QueryRowContext(ctx, query, args...), nil
}

func (l *LibSQLDB) Sync(ctx context.Context) error {
	_, err := l.DB.ExecContext(ctx, "SELECT sync()")
	return err
}

func CloudLogin(token string) error {
	return os.WriteFile(getCloudTokenPath(), []byte(token), 0600)
}

func CloudLogout() error {
	return os.Remove(getCloudTokenPath())
}

func GetCloudToken() string {
	data, _ := os.ReadFile(getCloudTokenPath())
	return string(data)
}

func getCloudTokenPath() string {
	home, _ := os.UserHomeDir()
	return home + "/.dataset-cli/.cloud_token"
}
