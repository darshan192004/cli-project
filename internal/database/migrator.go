package database

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"dataset-cli/internal/analyzer"
)

type ImportResult struct {
	SuccessCount int
	ErrorCount   int
	Errors       []ImportError
}

type ImportError struct {
	RecordIndex int
	Record      map[string]interface{}
	Error       string
}

type Migrator struct {
	db           Backend
	skipErrors   bool
	progressFunc func(current, total int)
	batchSize    int
}

type MigratorOption func(*Migrator)

func WithSkipErrors(skip bool) MigratorOption {
	return func(m *Migrator) {
		m.skipErrors = skip
	}
}

func WithProgressCallback(fn func(current, total int)) MigratorOption {
	return func(m *Migrator) {
		m.progressFunc = fn
	}
}

func WithBatchSize(size int) MigratorOption {
	return func(m *Migrator) {
		m.batchSize = size
	}
}

func NewMigrator(db Backend, opts ...MigratorOption) *Migrator {
	m := &Migrator{
		db:         db,
		skipErrors: false,
		batchSize:  1000,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func (m *Migrator) CreateTable(ctx context.Context, schema *analyzer.Schema) error {
	columns := make([]string, 0, len(schema.Columns))

	for _, col := range schema.Columns {
		colDef := fmt.Sprintf("%s %s", sanitizeColumnName(col.Name), mapToSQLType(col.Type, m.db.Type()))
		columns = append(columns, colDef)
	}

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", sanitizeTableName(schema.TableName), strings.Join(columns, ", "))

	_, err := m.db.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	return nil
}

func (m *Migrator) ImportData(ctx context.Context, schema *analyzer.Schema, records []map[string]interface{}) (*ImportResult, error) {
	if len(records) == 0 {
		return &ImportResult{}, nil
	}

	columns := make([]string, len(schema.Columns))
	for i, col := range schema.Columns {
		columns[i] = sanitizeColumnName(col.Name)
	}

	placeholder := "?"
	if m.db.Type() == BackendPostgres {
		placeholder = "$"
		for i := range columns {
			placeholder += fmt.Sprintf("%d", i+1)
			if i < len(columns)-1 {
				placeholder += ","
			}
		}
		placeholder = ""
		for i := range columns {
			if i > 0 {
				placeholder += ","
			}
			placeholder += fmt.Sprintf("$%d", i+1)
		}
	}

	placeholders := make([]string, len(columns))
	for i := range placeholders {
		if m.db.Type() == BackendPostgres {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
		} else {
			placeholders[i] = "?"
		}
	}

	insertQuery := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		sanitizeTableName(schema.TableName),
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	result := &ImportResult{
		Errors: make([]ImportError, 0),
	}

	totalRecords := len(records)
	successCount := int32(0)
	errorCount := int32(0)
	var mu sync.Mutex
	var wg sync.WaitGroup

	numWorkers := 4
	recordsPerWorker := (totalRecords + numWorkers - 1) / numWorkers

	for w := 0; w < numWorkers; w++ {
		start := w * recordsPerWorker
		end := start + recordsPerWorker
		if end > totalRecords {
			end = totalRecords
		}

		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()

			for i := start; i < end; i++ {
				record := records[i]
				values := make([]interface{}, len(schema.Columns))
				for j, col := range schema.Columns {
					val := record[col.Name]
					if str, ok := val.(string); ok {
						val = strings.TrimSpace(strings.ToValidUTF8(str, ""))
					}
					if val == "" {
						val = nil
					}
					values[j] = val
				}

				_, err := m.db.Exec(ctx, insertQuery, values...)
				if err != nil {
					if m.skipErrors {
						atomic.AddInt32(&errorCount, 1)
						mu.Lock()
						result.Errors = append(result.Errors, ImportError{
							RecordIndex: i,
							Record:      record,
							Error:       err.Error(),
						})
						mu.Unlock()
					} else {
						atomic.AddInt32(&errorCount, 1)
					}
				} else {
					atomic.AddInt32(&successCount, 1)
				}

				current := int(atomic.LoadInt32(&successCount) + atomic.LoadInt32(&errorCount))
				if m.progressFunc != nil && current%1000 == 0 {
					m.progressFunc(current, totalRecords)
				}
			}
		}(start, end)
	}

	wg.Wait()

	if m.progressFunc != nil {
		m.progressFunc(totalRecords, totalRecords)
	}

	result.SuccessCount = int(successCount)
	result.ErrorCount = int(errorCount)

	return result, nil
}

func sanitizeTableName(name string) string {
	return fmt.Sprintf(`"%s"`, name)
}

func sanitizeColumnName(name string) string {
	return fmt.Sprintf(`"%s"`, name)
}

func mapToSQLType(t analyzer.ColumnType, backend BackendType) string {
	switch t {
	case analyzer.TypeInteger:
		if backend == BackendSQLite || backend == BackendLibSQL {
			return "INTEGER"
		}
		return "INTEGER"
	case analyzer.TypeFloat:
		if backend == BackendSQLite || backend == BackendLibSQL {
			return "REAL"
		}
		return "DOUBLE PRECISION"
	case analyzer.TypeBoolean:
		if backend == BackendSQLite || backend == BackendLibSQL {
			return "INTEGER"
		}
		return "BOOLEAN"
	case analyzer.TypeTimestamp:
		if backend == BackendSQLite || backend == BackendLibSQL {
			return "TEXT"
		}
		return "TIMESTAMP"
	case analyzer.TypeDate:
		if backend == BackendSQLite || backend == BackendLibSQL {
			return "TEXT"
		}
		return "DATE"
	default:
		return "TEXT"
	}
}
