package analyzer

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"dataset-cli/internal/reader"
)

func sanitizeColumnName(name string) string {
	return reader.SanitizeColumnName(name)
}

type ColumnType int

const (
	TypeUnknown ColumnType = iota
	TypeString
	TypeInteger
	TypeFloat
	TypeBoolean
	TypeDate
	TypeTimestamp
)

func (t ColumnType) String() string {
	switch t {
	case TypeString:
		return "string"
	case TypeInteger:
		return "integer"
	case TypeFloat:
		return "float"
	case TypeBoolean:
		return "boolean"
	case TypeDate:
		return "date"
	case TypeTimestamp:
		return "timestamp"
	default:
		return "unknown"
	}
}

type Column struct {
	Name         string
	Type         ColumnType
	IsPrimaryKey bool
	SampleValues []interface{}
}

type Schema struct {
	TableName string
	Columns   []Column
}

type Analyzer struct {
	sampleSize int
}

func New() *Analyzer {
	return &Analyzer{sampleSize: 1000}
}

func (a *Analyzer) Analyze(filePath string) (*Schema, error) {
	filePath = strings.Trim(filePath, `"`)
	r, err := reader.GetReader(filePath)
	if err != nil {
		return nil, err
	}

	records, err := r.Read(filePath)
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("no records found in file")
	}

	headers, err := r.GetHeaders(filePath)
	if err != nil {
		return nil, err
	}

	columns := make([]Column, 0, len(headers))
	for _, header := range headers {
		colType := a.detectColumnType(records, header)
		sanitizedName := sanitizeColumnName(header)
		columns = append(columns, Column{
			Name:         sanitizedName,
			Type:         colType,
			IsPrimaryKey: false,
		})
	}

	tableName := sanitizeTableName(filepath.Base(filePath))

	return &Schema{
		TableName: tableName,
		Columns:   columns,
	}, nil
}

func (a *Analyzer) detectColumnType(records []map[string]interface{}, columnName string) ColumnType {
	sampleCount := 0
	if len(records) < a.sampleSize {
		sampleCount = len(records)
	} else {
		sampleCount = a.sampleSize
	}

	isInteger := true
	isFloat := true
	isBoolStrict := true
	isDate := true
	isTimestamp := true
	isNumericOnly := true
	hasDecimal := false
	nullCount := 0

	for i := 0; i < sampleCount; i++ {
		value := records[i][columnName]
		if value == nil || value == "" {
			nullCount++
			continue
		}

		strVal := fmt.Sprintf("%v", value)
		strVal = strings.TrimSpace(strVal)

		if isInteger && strVal != "" {
			if _, err := strconv.Atoi(strVal); err != nil {
				isInteger = false
			} else if strings.Contains(strVal, ".") {
				hasDecimal = true
			}
		}

		if isFloat && strVal != "" {
			if _, err := strconv.ParseFloat(strVal, 64); err != nil {
				isFloat = false
			} else if strings.Contains(strVal, ".") {
				hasDecimal = true
			}
		}

		if isBoolStrict && strVal != "" {
			lower := strings.ToLower(strVal)
			if lower != "true" && lower != "false" {
				isBoolStrict = false
			}
		}

		if isDate && strVal != "" {
			if _, err := time.Parse("2006-01-02", strVal); err != nil {
				isDate = false
			}
		}

		if isTimestamp && strVal != "" {
			if _, err := time.Parse(time.RFC3339, strVal); err != nil {
				isTimestamp = false
			}
		}

		if isNumericOnly && strVal != "" {
			if _, err := strconv.ParseFloat(strVal, 64); err != nil {
				isNumericOnly = false
			}
		}
	}

	nonNullCount := sampleCount - nullCount
	if nonNullCount == 0 {
		return TypeString
	}

	if !isNumericOnly {
		isInteger = false
		isFloat = false
	}

	if isTimestamp {
		return TypeTimestamp
	}
	if isDate {
		return TypeDate
	}
	if isFloat && hasDecimal {
		return TypeFloat
	}
	if isInteger {
		return TypeInteger
	}
	if isFloat && !hasDecimal {
		return TypeInteger
	}

	return TypeString
}

func sanitizeTableName(name string) string {
	name = strings.TrimSuffix(name, filepath.Ext(name))
	reg := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	name = reg.ReplaceAllString(name, "_")
	if len(name) > 63 {
		name = name[:63]
	}
	if name[0] >= '0' && name[0] <= '9' {
		name = "t_" + name
	}
	return strings.ToLower(name)
}
