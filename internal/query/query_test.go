package query

import (
	"testing"
)

func TestQueryBuilderFilter(t *testing.T) {
	q := New("users")

	query, args := q.Filter("age > 25")
	expectedQuery := `SELECT * FROM "users" WHERE age > 25`

	if query != expectedQuery {
		t.Errorf("Filter() query = %q; want %q", query, expectedQuery)
	}

	if len(args) != 0 {
		t.Errorf("Filter() args = %v; want empty", args)
	}
}

func TestQueryBuilderFilterWithArgs(t *testing.T) {
	q := New("users")

	query, _ := q.Filter("age > $1", 25)
	expectedQuery := `SELECT * FROM "users" WHERE age > $1`

	if query != expectedQuery {
		t.Errorf("Filter() query = %q; want %q", query, expectedQuery)
	}
}

func TestQueryBuilderTransform(t *testing.T) {
	q := New("users")

	query, _ := q.Transform([]string{"name", "email"}, "")
	expectedQuery := `SELECT "name", "email" FROM "users"`

	if query != expectedQuery {
		t.Errorf("Transform() query = %q; want %q", query, expectedQuery)
	}
}

func TestQueryBuilderTransformWithWhere(t *testing.T) {
	q := New("users")

	query, _ := q.Transform([]string{"name", "email"}, "active = true")
	expectedQuery := `SELECT "name", "email" FROM "users" WHERE active = true`

	if query != expectedQuery {
		t.Errorf("Transform() query = %q; want %q", query, expectedQuery)
	}
}

func TestQueryBuilderPaginate(t *testing.T) {
	q := New("users")

	query, _ := q.Paginate(10, 20, []string{})
	expectedQuery := `SELECT * FROM "users" ORDER BY 1 LIMIT 10 OFFSET 20`

	if query != expectedQuery {
		t.Errorf("Paginate() query = %q; want %q", query, expectedQuery)
	}
}

func TestQueryBuilderPaginateWithColumns(t *testing.T) {
	q := New("users")

	query, _ := q.Paginate(10, 0, []string{"id", "name"})
	expectedQuery := `SELECT "id", "name" FROM "users" ORDER BY 1 LIMIT 10 OFFSET 0`

	if query != expectedQuery {
		t.Errorf("Paginate() query = %q; want %q", query, expectedQuery)
	}
}

func TestQueryBuilderCount(t *testing.T) {
	q := New("users")

	query, _ := q.Count("")
	expectedQuery := `SELECT COUNT(*) FROM "users"`

	if query != expectedQuery {
		t.Errorf("Count() query = %q; want %q", query, expectedQuery)
	}
}

func TestQueryBuilderCountWithWhere(t *testing.T) {
	q := New("users")

	query, _ := q.Count("active = true")
	expectedQuery := `SELECT COUNT(*) FROM "users" WHERE active = true`

	if query != expectedQuery {
		t.Errorf("Count() query = %q; want %q", query, expectedQuery)
	}
}

func TestSanitizeColumnNames(t *testing.T) {
	tests := []struct {
		input    []string
		expected string
	}{
		{[]string{"name", "email"}, `"name", "email"`},
		{[]string{"id"}, `"id"`},
		{[]string{"first name", "last name"}, `"first name", "last name"`},
	}

	for _, tt := range tests {
		result := sanitizeColumnNames(tt.input)
		if result != tt.expected {
			t.Errorf("sanitizeColumnNames(%v) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}
