package errors

import (
	"fmt"
	"strings"
)

type AppError struct {
	Code    ErrorCode
	Message string
	Hint    string
	Err     error
}

type ErrorCode string

const (
	ErrFileNotFound  ErrorCode = "E001"
	ErrFileRead      ErrorCode = "E002"
	ErrFileFormat    ErrorCode = "E003"
	ErrDatabase      ErrorCode = "E004"
	ErrTableNotFound ErrorCode = "E005"
	ErrImportFailed  ErrorCode = "E006"
	ErrQueryFailed   ErrorCode = "E007"
	ErrValidation    ErrorCode = "E008"
	ErrConfig        ErrorCode = "E009"
)

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func (e *AppError) WithHint(hint string) *AppError {
	e.Hint = hint
	return e
}

func (e *AppError) WithError(err error) *AppError {
	e.Err = err
	return e
}

func (e *AppError) Print() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n  ✗ %s\n", e.Message))
	if e.Err != nil {
		sb.WriteString(fmt.Sprintf("    Error: %v\n", e.Err))
	}
	if e.Hint != "" {
		sb.WriteString(fmt.Sprintf("    💡 Hint: %s\n", e.Hint))
	}
	return sb.String()
}

func NewFileNotFoundError(path string) *AppError {
	return &AppError{
		Code:    ErrFileNotFound,
		Message: fmt.Sprintf("File not found: %s", path),
		Hint:    "Please check if the file path is correct and the file exists",
	}
}

func NewFileReadError(path string, err error) *AppError {
	return &AppError{
		Code:    ErrFileRead,
		Message: fmt.Sprintf("Failed to read file: %s", path),
		Err:     err,
		Hint:    "The file might be corrupted or have permission issues",
	}
}

func NewImportError(row int, column string, value string, err error) *AppError {
	return &AppError{
		Code:    ErrImportFailed,
		Message: fmt.Sprintf("Failed to import row %d: column '%s' has invalid value '%s'", row, column, value),
		Err:     err,
		Hint:    "Check if the data type matches the column schema",
	}
}

func NewDatabaseError(operation string, err error) *AppError {
	return &AppError{
		Code:    ErrDatabase,
		Message: fmt.Sprintf("Database error during %s", operation),
		Err:     err,
		Hint:    "Check if PostgreSQL is running and credentials are correct",
	}
}

func NewValidationError(field string, reason string) *AppError {
	return &AppError{
		Code:    ErrValidation,
		Message: fmt.Sprintf("Validation failed for '%s': %s", field, reason),
		Hint:    "Please check the input value",
	}
}

func NewTableNotFoundError(name string, suggestions []string) *AppError {
	hint := "Available tables: " + strings.Join(suggestions, ", ")
	return &AppError{
		Code:    ErrTableNotFound,
		Message: fmt.Sprintf("Table '%s' does not exist", name),
		Hint:    hint,
	}
}

func NewQueryError(query string, err error) *AppError {
	return &AppError{
		Code:    ErrQueryFailed,
		Message: fmt.Sprintf("Query failed: %s", err),
		Err:     err,
		Hint:    "Check if the query syntax is correct",
	}
}

func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

func GetErrorCode(err error) ErrorCode {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code
	}
	return ""
}
