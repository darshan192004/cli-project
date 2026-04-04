package mock

import (
	"context"
	"errors"
	"testing"
)

type MockBackend struct {
	Tables         []string
	TableSchemas   map[string][]TableSchema
	ExecuteResults []map[string]interface{}
	ExecuteError   error
	ExecResult     int64
	ExecError      error
}

type TableSchema struct {
	Name     string
	DataType string
}

func (m *MockBackend) Type() string { return "mock" }

func (m *MockBackend) TableExists(ctx context.Context, tableName string) (bool, error) {
	for _, t := range m.Tables {
		if t == tableName {
			return true, nil
		}
	}
	return false, nil
}

func (m *MockBackend) GetTableSchema(ctx context.Context, tableName string) ([]TableSchema, error) {
	if schema, ok := m.TableSchemas[tableName]; ok {
		return schema, nil
	}
	return nil, errors.New("table not found")
}

func (m *MockBackend) GetAllTables(ctx context.Context) ([]string, error) {
	return m.Tables, nil
}

func (m *MockBackend) Execute(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	if m.ExecuteError != nil {
		return nil, m.ExecuteError
	}
	return m.ExecuteResults, nil
}

func (m *MockBackend) Exec(ctx context.Context, query string, args ...interface{}) (int64, error) {
	if m.ExecError != nil {
		return 0, m.ExecError
	}
	return m.ExecResult, nil
}

func (m *MockBackend) Close() error { return nil }

func TestMockBackendTableExists(t *testing.T) {
	mock := &MockBackend{
		Tables: []string{"users", "products"},
	}

	ctx := context.Background()

	if exists, _ := mock.TableExists(ctx, "users"); !exists {
		t.Error("TableExists(users) = false, want true")
	}

	if exists, _ := mock.TableExists(ctx, "orders"); exists {
		t.Error("TableExists(orders) = true, want false")
	}
}

func TestMockBackendGetAllTables(t *testing.T) {
	mock := &MockBackend{
		Tables: []string{"users", "products", "orders"},
	}

	ctx := context.Background()
	tables, err := mock.GetAllTables(ctx)

	if err != nil {
		t.Fatalf("GetAllTables() error = %v", err)
	}

	if len(tables) != 3 {
		t.Errorf("len(tables) = %d, want 3", len(tables))
	}
}

func TestMockBackendExecute(t *testing.T) {
	mock := &MockBackend{
		ExecuteResults: []map[string]interface{}{
			{"id": 1, "name": "John"},
			{"id": 2, "name": "Jane"},
		},
	}

	ctx := context.Background()
	results, err := mock.Execute(ctx, "SELECT * FROM users")

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if len(results) != 2 {
		t.Errorf("len(results) = %d, want 2", len(results))
	}
}

func TestMockBackendExecuteError(t *testing.T) {
	mock := &MockBackend{
		ExecuteError: errors.New("query failed"),
	}

	ctx := context.Background()
	_, err := mock.Execute(ctx, "INVALID SQL")

	if err == nil {
		t.Error("Execute() should return error")
	}
}

func TestMockBackendExec(t *testing.T) {
	mock := &MockBackend{
		ExecResult: 5,
	}

	ctx := context.Background()
	count, err := mock.Exec(ctx, "DELETE FROM users")

	if err != nil {
		t.Fatalf("Exec() error = %v", err)
	}

	if count != 5 {
		t.Errorf("Exec() = %d, want 5", count)
	}
}

func TestMockBackendExecError(t *testing.T) {
	mock := &MockBackend{
		ExecError: errors.New("delete failed"),
	}

	ctx := context.Background()
	_, err := mock.Exec(ctx, "DELETE FROM users")

	if err == nil {
		t.Error("Exec() should return error")
	}
}
