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

	reader := csv.NewReader(file)
	reader.ReuseRecord = true

	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read headers: %w", err)
	}

	var cleanHeaders []string
	for _, h := range headers {
		h = strings.TrimSpace(h)
		if h != "" {
			cleanHeaders = append(cleanHeaders, h)
		}
	}
	headers = cleanHeaders
	numCols := len(headers)

	var records []map[string]interface{}
	for {
		record, err := reader.Read()
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
			row[header] = strings.TrimSpace(value)
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

	for i := range headers {
		headers[i] = strings.TrimSpace(headers[i])
	}

	return headers, nil
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

	return records, nil
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
