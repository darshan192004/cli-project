package config

import (
	"os"
	"testing"
)

func TestDatabaseConfigConnectionString(t *testing.T) {
	cfg := DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "user",
		Password: "password",
		DBName:   "testdb",
		SSLMode:  "disable",
	}

	expected := "postgres://user:password@localhost:5432/testdb?sslmode=disable"
	result := cfg.ConnectionString()

	if result != expected {
		t.Errorf("ConnectionString() = %q; want %q", result, expected)
	}
}

func TestLoadWithDefaults(t *testing.T) {
	originalHome := os.Getenv("HOME")
	os.Unsetenv("HOME")
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_USER")
	os.Unsetenv("DB_PASSWORD")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("DB_SSLMODE")

	cfg, err := Load()
	if err != nil {
		t.Logf("Load() returned error (expected if no config): %v", err)
	}

	if cfg != nil {
		if cfg.Database.Host != "localhost" {
			t.Errorf("Default host = %q; want localhost", cfg.Database.Host)
		}
		if cfg.Database.Port != 5432 {
			t.Errorf("Default port = %d; want 5432", cfg.Database.Port)
		}
		if cfg.Database.SSLMode != "disable" {
			t.Errorf("Default sslmode = %q; want disable", cfg.Database.SSLMode)
		}
	}

	if originalHome != "" {
		os.Setenv("HOME", originalHome)
	}
}

func TestDatabaseConfigConnectionStringWithSSL(t *testing.T) {
	cfg := DatabaseConfig{
		Host:     "db.example.com",
		Port:     5433,
		User:     "admin",
		Password: "pass123",
		DBName:   "production",
		SSLMode:  "require",
	}

	expected := "postgres://admin:pass123@db.example.com:5433/production?sslmode=require"
	result := cfg.ConnectionString()

	if result != expected {
		t.Errorf("ConnectionString() = %q, want %q", result, expected)
	}
}
