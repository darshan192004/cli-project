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
	return &Analyzer{sampleSize: 100}
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
		columns = append(columns, Column{
			Name:         header,
			Type:         colType,
			IsPrimaryKey: false,
		})
	}

	if len(columns) > 0 && columns[0].Type == TypeString {
		columns[0].IsPrimaryKey = true
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

	var isInteger, isFloat, isBool, isDate, isTimestamp bool
	nullCount := 0
	allNumeric := true
	hasDecimal := false

	for i := 0; i < sampleCount; i++ {
		value := records[i][columnName]
		if value == nil || value == "" {
			nullCount++
			continue
		}

		strVal := fmt.Sprintf("%v", value)

		if _, err := strconv.Atoi(strVal); err != nil {
			allNumeric = false
		} else {
			if strings.Contains(strVal, ".") {
				hasDecimal = true
			}
		}

		if isInteger && strVal != "" {
			if _, err := strconv.Atoi(strVal); err != nil {
				isInteger = false
			}
		} else if strVal != "" {
			if _, err := strconv.Atoi(strVal); err == nil {
				isInteger = true
			}
		}

		if isFloat && strVal != "" {
			if _, err := strconv.ParseFloat(strVal, 64); err != nil {
				isFloat = false
			}
		} else if strVal != "" {
			if _, err := strconv.ParseFloat(strVal, 64); err == nil {
				isFloat = true
			}
		}

		if isBool && strVal != "" {
			lower := strings.ToLower(strVal)
			if lower != "true" && lower != "false" && lower != "1" && lower != "0" {
				isBool = false
			}
		} else if strVal != "" {
			lower := strings.ToLower(strVal)
			if lower == "true" || lower == "false" || lower == "1" || lower == "0" {
				isBool = true
			}
		}

		if isDate && strVal != "" {
			if _, err := time.Parse("2006-01-02", strVal); err != nil {
				isDate = false
			}
		} else if strVal != "" {
			if _, err := time.Parse("2006-01-02", strVal); err == nil {
				isDate = true
			}
		}

		if isTimestamp && strVal != "" {
			if _, err := time.Parse(time.RFC3339, strVal); err != nil {
				isTimestamp = false
			}
		} else if strVal != "" {
			if _, err := time.Parse(time.RFC3339, strVal); err == nil {
				isTimestamp = true
			}
		}
	}

	nonNullCount := sampleCount - nullCount
	if nonNullCount == 0 {
		return TypeString
	}

	if !allNumeric {
		isInteger = false
		isFloat = false
	}

	if isFloat && !hasDecimal {
		return TypeInteger
	}
	if isFloat {
		return TypeFloat
	}
	if isInteger {
		return TypeInteger
	}
	if isBool {
		return TypeBoolean
	}
	if isTimestamp {
		return TypeTimestamp
	}
	if isDate {
		return TypeDate
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
