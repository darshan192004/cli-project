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
	db *database.DB
}

func NewExecutor(db *database.DB) *QueryExecutor {
	return &QueryExecutor{db: db}
}

func (e *QueryExecutor) Execute(query string, args ...interface{}) ([]map[string]interface{}, error) {
	ctx := context.Background()
	rows, err := e.db.Pool.Query(ctx, query, args...)
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

func (e *QueryExecutor) ExecuteAndCount(query string, args ...interface{}) (int64, []map[string]interface{}, error) {
	rows, err := e.db.Pool.Query(nil, query, args...)
	if err != nil {
		return 0, nil, err
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
			return 0, nil, err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}
		results = append(results, row)
	}

	var total int64
	countQuery := strings.Replace(query, "SELECT *", "SELECT COUNT(*)", 1)
	countQuery = strings.Replace(countQuery, "SELECT "+strings.Join(columns, ", "), "SELECT COUNT(*)", 1)

	// Simple count approach
	if len(results) > 0 {
		countQuery = fmt.Sprintf("SELECT COUNT(*) FROM (%s) as subquery", query)
		err = e.db.Pool.QueryRow(nil, countQuery, args...).Scan(&total)
		if err != nil {
			total = int64(len(results))
		}
	}

	return total, results, nil
}
