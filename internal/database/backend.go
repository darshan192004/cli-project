package database

import (
	"context"
	"fmt"
	"os"

	"dataset-cli/internal/config"
)

type BackendType string

const (
	BackendSQLite   BackendType = "sqlite"
	BackendPostgres BackendType = "postgres"
	BackendLibSQL   BackendType = "libsql"
)

type Backend interface {
	TableExists(ctx context.Context, tableName string) (bool, error)
	GetTableSchema(ctx context.Context, tableName string) ([]ColumnInfo, error)
	GetTableInfo(ctx context.Context, tableName string) (*TableInfo, error)
	GetAllTables(ctx context.Context) ([]string, error)
	Execute(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error)
	Exec(ctx context.Context, query string, args ...interface{}) (int64, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) (interface{}, error)
	Close() error
	Type() BackendType
}

func NewBackend(backend BackendType, cfg *config.Config) (Backend, error) {
	switch backend {
	case BackendSQLite:
		return ConnectSQLite("")
	case BackendPostgres:
		pg, err := Connect(&cfg.Database)
		if err != nil {
			return nil, err
		}
		return &PostgresBackend{DB: pg}, nil
	case BackendLibSQL:
		return ConnectLibSQL(os.Getenv("LIBSQL_URL"), os.Getenv("LIBSQL_AUTH_TOKEN"))
	default:
		return ConnectSQLite("")
	}
}

func ConnectFromFlags(usePostgres, useCloud bool) (Backend, error) {
	if useCloud {
		return NewBackend(BackendLibSQL, nil)
	}
	if usePostgres {
		cfg, err := config.Load()
		if err != nil {
			return nil, fmt.Errorf("failed to load config for PostgreSQL: %w", err)
		}
		return NewBackend(BackendPostgres, cfg)
	}
	return ConnectSQLite("")
}

type PostgresBackend struct {
	DB *DB
}

func (p *PostgresBackend) Type() BackendType { return BackendPostgres }

func (p *PostgresBackend) TableExists(ctx context.Context, tableName string) (bool, error) {
	return p.DB.TableExists(ctx, tableName)
}

func (p *PostgresBackend) GetTableSchema(ctx context.Context, tableName string) ([]ColumnInfo, error) {
	return p.DB.GetTableSchema(ctx, tableName)
}

func (p *PostgresBackend) GetTableInfo(ctx context.Context, tableName string) (*TableInfo, error) {
	return p.DB.GetTableInfo(ctx, tableName)
}

func (p *PostgresBackend) GetAllTables(ctx context.Context) ([]string, error) {
	return p.DB.GetAllTables(ctx)
}

func (p *PostgresBackend) Execute(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := p.DB.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fieldDescriptions := rows.FieldDescriptions()
	columns := make([]string, len(fieldDescriptions))
	for i, fd := range fieldDescriptions {
		columns[i] = string(fd.Name)
	}

	var results []map[string]interface{}
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}
		results = append(results, row)
	}

	return results, nil
}

func (p *PostgresBackend) Exec(ctx context.Context, query string, args ...interface{}) (int64, error) {
	result, err := p.DB.Pool.Exec(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

func (p *PostgresBackend) QueryRow(ctx context.Context, query string, args ...interface{}) (interface{}, error) {
	return p.DB.Pool.QueryRow(ctx, query, args...), nil
}

func (p *PostgresBackend) Close() error {
	p.DB.Close()
	return nil
}
