package validator

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"dataset-cli/internal/errors"
)

type ValidationResult struct {
	Valid      bool
	Errors     []*errors.AppError
	Warnings   []string
	FileInfo   *FileInfo
	SchemaInfo *SchemaInfo
}

type FileInfo struct {
	Path      string
	Name      string
	Extension string
	SizeBytes int64
	LineCount int
	Encoding  string
}

type SchemaInfo struct {
	Columns         []ColumnInfo
	TotalRows       int
	EstimatedSizeMB float64
}

type ColumnInfo struct {
	Name         string
	SampleValues []string
	NullCount    int
	IsMixedType  bool
}

func ValidateFile(path string) *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   make([]*errors.AppError, 0),
		Warnings: make([]string, 0),
	}

	path = strings.Trim(path, "\"")

	info, err := os.Stat(path)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, errors.NewFileNotFoundError(path))
		return result
	}

	if info.IsDir() {
		result.Valid = false
		result.Errors = append(result.Errors, errors.NewValidationError("path", "is a directory, not a file"))
		return result
	}

	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".csv" && ext != ".json" {
		result.Valid = false
		result.Errors = append(result.Errors, errors.NewValidationError("file format", fmt.Sprintf("'%s' is not supported. Use .csv or .json", ext)))
		return result
	}

	file, err := os.Open(path)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, errors.NewFileReadError(path, err))
		return result
	}
	defer func() { _ = file.Close() }()

	result.FileInfo = &FileInfo{
		Path:      path,
		Name:      filepath.Base(path),
		Extension: ext,
		SizeBytes: info.Size(),
	}

	if ext == ".csv" {
		result.validateCSV(file)
	} else {
		result.validateJSON(file)
	}

	if result.FileInfo.SizeBytes > 100*1024*1024 {
		mb := float64(result.FileInfo.SizeBytes) / (1024 * 1024)
		result.Warnings = append(result.Warnings, fmt.Sprintf("Large file detected (%.1f MB). Import may take longer.", mb))
	}

	if result.FileInfo.LineCount > 100000 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("File has %d rows. Consider using --skip-errors for faster import.", result.FileInfo.LineCount))
	}

	return result
}

func (r *ValidationResult) validateCSV(file *os.File) {
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true

	headers, err := reader.Read()
	if err != nil {
		r.Valid = false
		r.Errors = append(r.Errors, errors.NewFileReadError(r.FileInfo.Path, err))
		return
	}

	headerCount := len(headers)
	hasEmptyHeader := false
	for i, h := range headers {
		h = strings.TrimSpace(h)
		if h == "" {
			hasEmptyHeader = true
			r.Warnings = append(r.Warnings, fmt.Sprintf("Empty header found at position %d", i+1))
		}
		headers[i] = h
	}

	if hasEmptyHeader {
		r.Warnings = append(r.Warnings, "Some columns have empty headers and will be named automatically")
	}

	lineCount := 1
	rowCount := 0
	colCounts := make(map[int]int)
	minCols := 0
	maxCols := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		lineCount++
		rowCount++

		colCount := len(record)
		if minCols == 0 || colCount < minCols {
			minCols = colCount
		}
		if colCount > maxCols {
			maxCols = colCount
		}

		colCounts[colCount]++

		if lineCount > 1000 {
			break
		}
	}

	r.FileInfo.LineCount = lineCount

	if minCols != maxCols && maxCols > 0 {
		r.Warnings = append(r.Warnings, fmt.Sprintf("Inconsistent column count: rows have between %d and %d columns", minCols, maxCols))
	}

	mostCommonCols := 0
	maxCount := 0
	for cols, count := range colCounts {
		if count > maxCount {
			maxCount = count
			mostCommonCols = cols
		}
	}

	if headerCount != mostCommonCols {
		r.Warnings = append(r.Warnings, fmt.Sprintf("Header has %d columns but most rows have %d columns", headerCount, mostCommonCols))
	}

	r.SchemaInfo = &SchemaInfo{
		Columns:         make([]ColumnInfo, 0),
		TotalRows:       rowCount,
		EstimatedSizeMB: float64(r.FileInfo.SizeBytes) / (1024 * 1024),
	}

	for i, h := range headers {
		r.SchemaInfo.Columns = append(r.SchemaInfo.Columns, ColumnInfo{
			Name:         h,
			SampleValues: make([]string, 0),
		})
		if i >= mostCommonCols {
			r.SchemaInfo.Columns[i].IsMixedType = true
		}
	}
}

func (r *ValidationResult) validateJSON(file *os.File) {
	data, err := io.ReadAll(file)
	if err != nil {
		r.Valid = false
		r.Errors = append(r.Errors, errors.NewFileReadError(r.FileInfo.Path, err))
		return
	}

	r.FileInfo.LineCount = strings.Count(string(data), "\n") + 1
	r.FileInfo.Encoding = "UTF-8"

	r.SchemaInfo = &SchemaInfo{
		Columns:         make([]ColumnInfo, 0),
		TotalRows:       1,
		EstimatedSizeMB: float64(r.FileInfo.SizeBytes) / (1024 * 1024),
	}
}

func PrintValidationResult(result *ValidationResult) {
	if result.Valid {
		fmt.Println("\n  ✓ File validation passed")
	} else {
		fmt.Println("\n  ✗ File validation failed")
	}

	if result.FileInfo != nil {
		fmt.Printf("  • File: %s\n", result.FileInfo.Name)
		fmt.Printf("  • Size: %.2f MB\n", float64(result.FileInfo.SizeBytes)/(1024*1024))
		fmt.Printf("  • Type: %s\n", strings.ToUpper(result.FileInfo.Extension[1:]))
	}

	if result.SchemaInfo != nil {
		fmt.Printf("  • Columns: %d\n", len(result.SchemaInfo.Columns))
		fmt.Printf("  • Sample rows: %d\n", result.SchemaInfo.TotalRows)
	}

	for _, err := range result.Errors {
		fmt.Print(err.Print())
	}

	if len(result.Warnings) > 0 {
		fmt.Println("\n  ⚠ Warnings:")
		for _, w := range result.Warnings {
			fmt.Printf("    • %s\n", w)
		}
	}
}
