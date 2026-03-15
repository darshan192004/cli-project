package reader

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCSVReader(t *testing.T) {
	tmpDir := t.TempDir()

	csvContent := "id,name,age\n1,John,25\n2,Jane,30"

	csvPath := filepath.Join(tmpDir, "test.csv")
	err := os.WriteFile(csvPath, []byte(csvContent), 0644)
	if err != nil {
		t.Fatalf("failed to create test CSV: %v", err)
	}

	reader := &CSVReader{}
	records, err := reader.Read(csvPath)
	if err != nil {
		t.Fatalf("CSVReader.Read() error = %v", err)
	}

	if len(records) != 2 {
		t.Errorf("len(records) = %d; want 2", len(records))
	}

	if len(records[0]) != 3 {
		t.Errorf("len(records[0]) = %d; want 3", len(records[0]))
	}
}

func TestCSVReaderGetHeaders(t *testing.T) {
	tmpDir := t.TempDir()

	csvContent := `id,name,age
1,John,25`

	csvPath := filepath.Join(tmpDir, "test.csv")
	err := os.WriteFile(csvPath, []byte(csvContent), 0644)
	if err != nil {
		t.Fatalf("failed to create test CSV: %v", err)
	}

	reader := &CSVReader{}
	headers, err := reader.GetHeaders(csvPath)
	if err != nil {
		t.Fatalf("CSVReader.GetHeaders() error = %v", err)
	}

	expected := []string{"id", "name", "age"}
	if len(headers) != len(expected) {
		t.Errorf("len(headers) = %d; want %d", len(headers), len(expected))
	}
}

func TestJSONReader(t *testing.T) {
	tmpDir := t.TempDir()

	jsonContent := `[{"id":1,"name":"John","age":25},{"id":2,"name":"Jane","age":30}]`

	jsonPath := filepath.Join(tmpDir, "test.json")
	err := os.WriteFile(jsonPath, []byte(jsonContent), 0644)
	if err != nil {
		t.Fatalf("failed to create test JSON: %v", err)
	}

	reader := &JSONReader{}
	records, err := reader.Read(jsonPath)
	if err != nil {
		t.Fatalf("JSONReader.Read() error = %v", err)
	}

	if len(records) != 2 {
		t.Errorf("len(records) = %d; want 2", len(records))
	}

	if records[0]["name"] != "John" {
		t.Errorf("records[0][name] = %v; want John", records[0]["name"])
	}
}

func TestJSONReaderSingleObject(t *testing.T) {
	tmpDir := t.TempDir()

	jsonContent := `{"id":1,"name":"John","age":25}`

	jsonPath := filepath.Join(tmpDir, "test.json")
	err := os.WriteFile(jsonPath, []byte(jsonContent), 0644)
	if err != nil {
		t.Fatalf("failed to create test JSON: %v", err)
	}

	reader := &JSONReader{}
	records, err := reader.Read(jsonPath)
	if err != nil {
		t.Fatalf("JSONReader.Read() error = %v", err)
	}

	if len(records) != 1 {
		t.Errorf("len(records) = %d; want 1", len(records))
	}
}

func TestGetReader(t *testing.T) {
	tests := []struct {
		ext      string
		wantErr  bool
	}{
		{".csv", false},
		{".json", false},
		{".xml", true},
		{".txt", true},
	}

	for _, tt := range tests {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "test"+tt.ext)
		os.WriteFile(path, []byte("test"), 0644)

		_, err := GetReader(path)
		if (err != nil) != tt.wantErr {
			t.Errorf("GetReader() error = %v, wantErr %v", err, tt.wantErr)
		}
	}
}
