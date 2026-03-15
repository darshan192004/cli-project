package database

import (
	"context"
	"fmt"

	"dataset-cli/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func Connect(cfg *config.DatabaseConfig) (*DB, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{Pool: pool}, nil
}

func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

func (db *DB) TableExists(ctx context.Context, tableName string) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = $1
		)`
	err := db.Pool.QueryRow(ctx, query, tableName).Scan(&exists)
	return exists, err
}

func (db *DB) GetTableSchema(ctx context.Context, tableName string) ([]ColumnInfo, error) {
	query := `
		SELECT column_name, data_type, is_nullable, column_default
		FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = $1
		ORDER BY ordinal_position`

	rows, err := db.Pool.Query(ctx, query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		if err := rows.Scan(&col.Name, &col.DataType, &col.IsNullable, &col.Default); err != nil {
			return nil, err
		}
		columns = append(columns, col)
	}
	return columns, nil
}

type ColumnInfo struct {
	Name       string
	DataType   string
	IsNullable string
	Default    *string
}

type TableInfo struct {
	Name    string
	Columns []ColumnInfo
	Count   int64
}

func (db *DB) GetTableInfo(ctx context.Context, tableName string) (*TableInfo, error) {
	exists, err := db.TableExists(ctx, tableName)
	if err != nil || !exists {
		return nil, fmt.Errorf("table '%s' does not exist", tableName)
	}

	columns, err := db.GetTableSchema(ctx, tableName)
	if err != nil {
		return nil, err
	}

	var count int64
	err = db.Pool.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&count)
	if err != nil {
		return nil, err
	}

	return &TableInfo{
		Name:    tableName,
		Columns: columns,
		Count:   count,
	}, nil
}

func (db *DB) GetAllTables(ctx context.Context) ([]string, error) {
	query := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public'
		ORDER BY table_name`

	rows, err := db.Pool.Query(ctx, query)
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
