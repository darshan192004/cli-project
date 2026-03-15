package database

import (
	"context"
	"fmt"
	"strings"

	"dataset-cli/internal/analyzer"
	"github.com/jackc/pgx/v5"
)

type Migrator struct {
	db *DB
}

func NewMigrator(db *DB) *Migrator {
	return &Migrator{db: db}
}

func (m *Migrator) CreateTable(ctx context.Context, schema *analyzer.Schema) error {
	columns := make([]string, 0, len(schema.Columns))
	pkColumns := make([]string, 0)

	for _, col := range schema.Columns {
		colDef := fmt.Sprintf("%s %s", sanitizeColumnName(col.Name), mapToPostgresType(col.Type))
		if col.IsPrimaryKey {
			pkColumns = append(pkColumns, sanitizeColumnName(col.Name))
		}
		columns = append(columns, colDef)
	}

	if len(pkColumns) > 0 {
		columns = append(columns, fmt.Sprintf("PRIMARY KEY (%s)", strings.Join(pkColumns, ", ")))
	}

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", sanitizeTableName(schema.TableName), strings.Join(columns, ", "))

	_, err := m.db.Pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	return nil
}

func (m *Migrator) ImportData(ctx context.Context, schema *analyzer.Schema, records []map[string]interface{}) error {
	if len(records) == 0 {
		return nil
	}

	columns := make([]string, len(schema.Columns))
	for i, col := range schema.Columns {
		columns[i] = sanitizeColumnName(col.Name)
	}

	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	batch := &pgx.Batch{}
	for _, record := range records {
		values := make([]interface{}, len(schema.Columns))
		for i, col := range schema.Columns {
			val := record[col.Name]
			if str, ok := val.(string); ok {
				val = strings.ToValidUTF8(str, "")
			}
			values[i] = val
		}
		query := fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s)",
			sanitizeTableName(schema.TableName),
			strings.Join(columns, ", "),
			strings.Join(placeholders, ", "),
		)
		batch.Queue(query, values...)
	}

	br := m.db.Pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < batch.Len(); i++ {
		_, err := br.Exec()
		if err != nil {
			return fmt.Errorf("failed to insert record %d: %w", i, err)
		}
	}

	return nil
}

func sanitizeTableName(name string) string {
	return fmt.Sprintf(`"%s"`, name)
}

func sanitizeColumnName(name string) string {
	return fmt.Sprintf(`"%s"`, name)
}

func mapToPostgresType(t analyzer.ColumnType) string {
	switch t {
	case analyzer.TypeInteger:
		return "INTEGER"
	case analyzer.TypeFloat:
		return "DOUBLE PRECISION"
	case analyzer.TypeBoolean:
		return "BOOLEAN"
	case analyzer.TypeTimestamp:
		return "TIMESTAMP"
	case analyzer.TypeDate:
		return "DATE"
	default:
		return "TEXT"
	}
}
