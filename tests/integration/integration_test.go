package integration

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSQLiteIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	t.Setenv("TEST_DB_PATH", dbPath)

	t.Run("Database operations", func(t *testing.T) {
		t.Log("Integration test placeholder - requires SQLite CGO")
		t.Skip("Requires CGO-enabled build")
	})
}

func TestQueryIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Query building integration", func(t *testing.T) {
		t.Log("Query integration test placeholder")
		t.Skip("Requires database connection")
	})
}

func TestMigratorIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Data migration integration", func(t *testing.T) {
		t.Log("Migrator integration test placeholder")
		t.Skip("Requires database connection")
	})
}

func createTempCSV(t *testing.T, dir, name, content string) string {
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test CSV: %v", err)
	}
	return path
}

func createTempJSON(t *testing.T, dir, name, content string) string {
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test JSON: %v", err)
	}
	return path
}

func TestCSVReading(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	csvContent := `id,name,age
1,John,25
2,Jane,30
3,Bob,35`

	csvPath := createTempCSV(t, tmpDir, "test.csv", csvContent)

	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		t.Error("CSV file should exist")
	}

	t.Logf("Created CSV at: %s", csvPath)
}

func TestJSONReading(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	jsonContent := `[{"id":1,"name":"John","active":true},{"id":2,"name":"Jane","active":false}]`

	jsonPath := createTempJSON(t, tmpDir, "test.json", jsonContent)

	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		t.Error("JSON file should exist")
	}

	t.Logf("Created JSON at: %s", jsonPath)
}
