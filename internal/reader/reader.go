package reader

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Reader interface {
	Read(path string) ([]map[string]interface{}, error)
	GetHeaders(path string) ([]string, error)
}

type CSVReader struct{}

func (r *CSVReader) Read(path string) ([]map[string]interface{}, error) {
	path = strings.Trim(path, `"`)
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	csvReader := csv.NewReader(file)
	csvReader.ReuseRecord = true

	headers, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read headers: %w", err)
	}

	var cleanHeaders []string
	for _, h := range headers {
		h = strings.TrimSpace(h)
		if h != "" {
			cleanHeaders = append(cleanHeaders, SanitizeColumnName(h))
		}
	}
	headers = cleanHeaders
	numCols := len(headers)

	var records []map[string]interface{}
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read record: %w", err)
		}

		row := make(map[string]interface{})
		for i, value := range record {
			if i >= numCols {
				break
			}
			header := headers[i]
			trimmedValue := strings.TrimSpace(value)
			if trimmedValue == "" {
				row[header] = nil
			} else {
				row[header] = trimmedValue
			}
		}
		records = append(records, row)
	}

	return records, nil
}

func (r *CSVReader) GetHeaders(path string) ([]string, error) {
	path = strings.Trim(path, `"`)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	headers, err := reader.Read()
	if err != nil {
		return nil, err
	}

	var sanitizedHeaders []string
	for _, h := range headers {
		h = strings.TrimSpace(h)
		if h != "" {
			sanitizedHeaders = append(sanitizedHeaders, SanitizeColumnName(h))
		}
	}

	return sanitizedHeaders, nil
}

type JSONReader struct{}

func (r *JSONReader) Read(path string) ([]map[string]interface{}, error) {
	path = strings.Trim(path, `"`)
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var records []map[string]interface{}
	if err := json.Unmarshal(data, &records); err != nil {
		var singleRecord map[string]interface{}
		if err := json.Unmarshal(data, &singleRecord); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
		records = []map[string]interface{}{singleRecord}
	}

	sanitizedRecords := make([]map[string]interface{}, len(records))
	for i, record := range records {
		sanitized := make(map[string]interface{})
		for key, value := range record {
			sanitizedKey := SanitizeColumnName(key)
			if value == nil || value == "" {
				sanitized[sanitizedKey] = nil
			} else {
				sanitized[sanitizedKey] = value
			}
		}
		sanitizedRecords[i] = sanitized
	}

	return sanitizedRecords, nil
}

func (r *JSONReader) GetHeaders(path string) ([]string, error) {
	records, err := r.Read(path)
	if err != nil || len(records) == 0 {
		return nil, err
	}

	headers := make([]string, 0)
	for key := range records[0] {
		headers = append(headers, key)
	}
	return headers, nil
}

func GetReader(path string) (Reader, error) {
	path = strings.Trim(path, `"`)
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".csv":
		return &CSVReader{}, nil
	case ".json":
		return &JSONReader{}, nil
	default:
		return nil, fmt.Errorf("unsupported file format: %s (supported: csv, json)", ext)
	}
}
