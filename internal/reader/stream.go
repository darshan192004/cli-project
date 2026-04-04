package reader

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type StreamConfig struct {
	BatchSize  int
	Callback   func(batch []map[string]interface{}, batchNum int) error
	BufferSize int
}

type StreamReader struct {
	filePath string
	config   *StreamConfig
}

func NewStreamReader(filePath string, config *StreamConfig) *StreamReader {
	if config == nil {
		config = &StreamConfig{
			BatchSize:  1000,
			BufferSize: 10000,
		}
	}
	if config.BatchSize == 0 {
		config.BatchSize = 1000
	}
	return &StreamReader{
		filePath: strings.Trim(filePath, "\""),
		config:   config,
	}
}

func (sr *StreamReader) Stream() error {
	ext := strings.ToLower(filepath.Ext(sr.filePath))
	switch ext {
	case ".csv":
		return sr.streamCSV()
	case ".json":
		return sr.streamJSON()
	default:
		return fmt.Errorf("unsupported file format: %s", ext)
	}
}

func (sr *StreamReader) streamCSV() error {
	file, err := os.Open(sr.filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	reader := csv.NewReader(file)
	reader.ReuseRecord = true
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true

	headers, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read headers: %w", err)
	}

	headers = sr.cleanHeaders(headers)

	batch := make([]map[string]interface{}, 0, sr.config.BatchSize)
	batchNum := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			if len(batch) > 0 {
				batchNum++
				if err := sr.config.Callback(batch, batchNum); err != nil {
					return err
				}
			}
			break
		}
		if err != nil {
			continue
		}

		row := sr.csvRowToMap(record, headers)
		batch = append(batch, row)

		if len(batch) >= sr.config.BatchSize {
			batchNum++
			if err := sr.config.Callback(batch, batchNum); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}

	return nil
}

func (sr *StreamReader) streamJSON() error {
	file, err := os.Open(sr.filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var records []map[string]interface{}
	if err := json.Unmarshal(data, &records); err != nil {
		var singleRecord map[string]interface{}
		if err := json.Unmarshal(data, &singleRecord); err != nil {
			return fmt.Errorf("failed to parse JSON: %w", err)
		}
		records = []map[string]interface{}{singleRecord}
	}

	batch := make([]map[string]interface{}, 0, sr.config.BatchSize)
	batchNum := 0

	for _, record := range records {
		sanitized := sr.sanitizeRecord(record)
		batch = append(batch, sanitized)

		if len(batch) >= sr.config.BatchSize {
			batchNum++
			if err := sr.config.Callback(batch, batchNum); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}

	if len(batch) > 0 {
		batchNum++
		if err := sr.config.Callback(batch, batchNum); err != nil {
			return err
		}
	}

	return nil
}

func (sr *StreamReader) cleanHeaders(headers []string) []string {
	clean := make([]string, 0, len(headers))
	for _, h := range headers {
		h = strings.TrimSpace(h)
		if h != "" {
			clean = append(clean, SanitizeColumnName(h))
		}
	}
	return clean
}

func (sr *StreamReader) csvRowToMap(record []string, headers []string) map[string]interface{} {
	row := make(map[string]interface{})
	for i, value := range record {
		if i >= len(headers) {
			break
		}
		header := headers[i]
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			row[header] = nil
		} else {
			row[header] = trimmed
		}
	}
	return row
}

func (sr *StreamReader) sanitizeRecord(record map[string]interface{}) map[string]interface{} {
	sanitized := make(map[string]interface{})
	for key, value := range record {
		sanitizedKey := SanitizeColumnName(key)
		if value == nil || value == "" {
			sanitized[sanitizedKey] = nil
		} else {
			sanitized[sanitizedKey] = value
		}
	}
	return sanitized
}

func CountLines(filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer func() { _ = file.Close() }()

	lineCount := 0
	buf := make([]byte, 32*1024)
	for {
		n, err := file.Read(buf)
		if n > 0 {
			for i := 0; i < n; i++ {
				if buf[i] == '\n' {
					lineCount++
				}
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
	}

	return lineCount, nil
}

func SanitizeColumnName(name string) string {
	reg := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	name = reg.ReplaceAllString(name, "_")
	name = strings.Trim(name, "_")
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ToLower(name)
	if name == "" {
		name = "column"
	}
	if len(name) > 0 && name[0] >= '0' && name[0] <= '9' {
		name = "c_" + name
	}
	return name
}
