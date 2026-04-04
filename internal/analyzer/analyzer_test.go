package analyzer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSanitizeTableName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"users", "users"},
		{"my table", "my_table"},
		{"table-123", "table_123"},
		{"123table", "t_123table"},
	}

	for _, tt := range tests {
		result := sanitizeTableName(tt.input)
		if result != tt.expected {
			t.Errorf("sanitizeTableName(%q) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}

func TestColumnTypeString(t *testing.T) {
	tests := []struct {
		ct       ColumnType
		expected string
	}{
		{TypeString, "string"},
		{TypeInteger, "integer"},
		{TypeFloat, "float"},
		{TypeBoolean, "boolean"},
		{TypeDate, "date"},
		{TypeTimestamp, "timestamp"},
		{TypeUnknown, "unknown"},
	}

	for _, tt := range tests {
		result := tt.ct.String()
		if result != tt.expected {
			t.Errorf("ColumnType.String() = %q; want %q", result, tt.expected)
		}
	}
}

func TestAnalyzerAnalyze(t *testing.T) {
	tmpDir := t.TempDir()

	csvContent := `id,name,age,city
1,John,25,NYC
2,Jane,30,LA
3,Bob,35,Chicago`

	csvPath := filepath.Join(tmpDir, "test.csv")
	err := os.WriteFile(csvPath, []byte(csvContent), 0644)
	if err != nil {
		t.Fatalf("failed to create test CSV: %v", err)
	}

	analyzer := New()
	schema, err := analyzer.Analyze(csvPath)
	if err != nil {
		t.Fatalf("Analyzer.Analyze() error = %v", err)
	}

	if schema.TableName != "test" {
		t.Errorf("TableName = %q; want %q", schema.TableName, "test")
	}

	if len(schema.Columns) != 4 {
		t.Errorf("len(Columns) = %d; want 4", len(schema.Columns))
	}

	expectedCols := []string{"id", "name", "age", "city"}
	for i, col := range schema.Columns {
		if col.Name != expectedCols[i] {
			t.Errorf("Column[%d].Name = %q; want %q", i, col.Name, expectedCols[i])
		}
	}
}

func TestAnalyzerAnalyzeJSON(t *testing.T) {
	tmpDir := t.TempDir()

	jsonContent := `[{"id":1,"name":"John","active":true},{"id":2,"name":"Jane","active":false}]`

	jsonPath := filepath.Join(tmpDir, "test.json")
	err := os.WriteFile(jsonPath, []byte(jsonContent), 0644)
	if err != nil {
		t.Fatalf("failed to create test JSON: %v", err)
	}

	analyzer := New()
	schema, err := analyzer.Analyze(jsonPath)
	if err != nil {
		t.Fatalf("Analyzer.Analyze() error = %v", err)
	}

	if schema.TableName != "test" {
		t.Errorf("TableName = %q; want %q", schema.TableName, "test")
	}

	if len(schema.Columns) != 3 {
		t.Errorf("len(Columns) = %d; want 3", len(schema.Columns))
	}
}

func TestSanitizeColumnName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"ITEM CODE", "item_code"},
		{"First-Name", "first_name"},
		{"Name123", "name123"},
		{"user@email", "user_email"},
		{"  spaces  ", "spaces"},
		{"UPPERCASE", "uppercase"},
		{"MixedCase", "mixedcase"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeColumnName(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeColumnName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
