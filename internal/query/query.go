package query

import (
	"context"
	"fmt"
	"strings"

	"dataset-cli/internal/database"
)

type QueryBuilder struct {
	tableName string
}

func New(tableName string) *QueryBuilder {
	return &QueryBuilder{tableName: tableName}
}

func (q *QueryBuilder) Filter(where string, args ...interface{}) (string, []interface{}) {
	query := fmt.Sprintf("SELECT * FROM %s", sanitizeTableName(q.tableName))

	if where != "" {
		query += " WHERE " + where
	}

	return query, args
}

func (q *QueryBuilder) Transform(columns []string, where string) (string, []interface{}) {
	cols := "*"
	if len(columns) > 0 {
		cols = sanitizeColumnNames(columns)
	}

	query := fmt.Sprintf("SELECT %s FROM %s", cols, sanitizeTableName(q.tableName))

	if where != "" {
		query += " WHERE " + where
	}

	return query, nil
}

func (q *QueryBuilder) Paginate(limit, offset int, columns []string) (string, []interface{}) {
	cols := "*"
	if len(columns) > 0 {
		cols = sanitizeColumnNames(columns)
	}

	query := fmt.Sprintf("SELECT %s FROM %s ORDER BY 1 LIMIT %d OFFSET %d",
		cols, sanitizeTableName(q.tableName), limit, offset)

	return query, nil
}

func (q *QueryBuilder) Select(columns []string, where string, limit, offset int) (string, []interface{}) {
	cols := "*"
	if len(columns) > 0 {
		cols = sanitizeColumnNames(columns)
	}

	query := fmt.Sprintf("SELECT %s FROM %s", cols, sanitizeTableName(q.tableName))

	if where != "" {
		query += " WHERE " + where
	}

	if limit > 0 {
		query += fmt.Sprintf(" ORDER BY 1 LIMIT %d", limit)
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}

	return query, nil
}

func (q *QueryBuilder) Count(where string) (string, []interface{}) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", sanitizeTableName(q.tableName))

	if where != "" {
		query += " WHERE " + where
	}

	return query, nil
}

func sanitizeTableName(name string) string {
	return fmt.Sprintf(`"%s"`, name)
}

func sanitizeColumnNames(columns []string) string {
	sanitized := make([]string, len(columns))
	for i, col := range columns {
		sanitized[i] = fmt.Sprintf(`"%s"`, col)
	}
	return strings.Join(sanitized, ", ")
}

type QueryExecutor struct {
	db database.Backend
}

func NewExecutor(db database.Backend) *QueryExecutor {
	return &QueryExecutor{db: db}
}

func (e *QueryExecutor) Execute(query string, args ...interface{}) ([]map[string]interface{}, error) {
	ctx := context.Background()
	return e.db.Execute(ctx, query, args...)
}

func (e *QueryExecutor) ExecuteAndCount(query string, args ...interface{}) (int64, []map[string]interface{}, error) {
	ctx := context.Background()
	results, err := e.db.Execute(ctx, query, args...)
	if err != nil {
		return 0, nil, err
	}

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM (%s) as subquery", query)
	row, _ := e.db.QueryRow(ctx, countQuery, args...)
	row.(interface{ Scan(...interface{}) error }).Scan(&total)

	return total, results, nil
}
