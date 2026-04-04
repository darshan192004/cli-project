package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
)

type OutputFormat string

const (
	FormatTable    OutputFormat = "table"
	FormatJSON     OutputFormat = "json"
	FormatCSV      OutputFormat = "csv"
	FormatMarkdown OutputFormat = "md"
	FormatPretty   OutputFormat = "pretty"
)

type Formatter interface {
	Format(results []map[string]interface{}) string
}

type JSONFormatter struct {
	Pretty bool
}

func NewJSONFormatter(pretty bool) *JSONFormatter {
	return &JSONFormatter{Pretty: pretty}
}

func (f *JSONFormatter) Format(results []map[string]interface{}) string {
	if f.Pretty {
		data, _ := json.MarshalIndent(results, "", "  ")
		return string(data)
	}
	data, _ := json.Marshal(results)
	return string(data)
}

type CSVFormatter struct{}

func NewCSVFormatter() *CSVFormatter {
	return &CSVFormatter{}
}

func (f *CSVFormatter) Format(results []map[string]interface{}) string {
	if len(results) == 0 {
		return ""
	}

	var sb strings.Builder

	cols := make([]string, 0)
	for key := range results[0] {
		cols = append(cols, key)
	}

	sb.WriteString(strings.Join(cols, ","))
	sb.WriteString("\n")

	for _, row := range results {
		var vals []string
		for _, col := range cols {
			val := ""
			if v, ok := row[col]; ok && v != nil {
				val = fmt.Sprintf("%v", v)
				if strings.Contains(val, ",") || strings.Contains(val, "\"") || strings.Contains(val, "\n") {
					val = "\"" + strings.ReplaceAll(val, "\"", "\"\"") + "\""
				}
			}
			vals = append(vals, val)
		}
		sb.WriteString(strings.Join(vals, ","))
		sb.WriteString("\n")
	}

	return sb.String()
}

type MarkdownFormatter struct{}

func NewMarkdownFormatter() *MarkdownFormatter {
	return &MarkdownFormatter{}
}

func (f *MarkdownFormatter) Format(results []map[string]interface{}) string {
	if len(results) == 0 {
		return "No results"
	}

	var sb strings.Builder

	cols := make([]string, 0)
	for key := range results[0] {
		cols = append(cols, key)
	}

	sb.WriteString("| ")
	sb.WriteString(strings.Join(cols, " | "))
	sb.WriteString(" |\n")

	sb.WriteString("|")
	for i := 0; i < len(cols); i++ {
		sb.WriteString(" --- |")
	}
	sb.WriteString("\n")

	for _, row := range results {
		sb.WriteString("| ")
		for i, col := range cols {
			val := ""
			if v, ok := row[col]; ok && v != nil {
				val = fmt.Sprintf("%v", v)
				val = strings.ReplaceAll(val, "\n", " ")
				val = strings.ReplaceAll(val, "|", "\\|")
				if len(val) > 50 {
					val = val[:47] + "..."
				}
			}
			sb.WriteString(val)
			if i < len(cols)-1 {
				sb.WriteString(" | ")
			}
		}
		sb.WriteString(" |\n")
	}

	return sb.String()
}

type PrettyFormatter struct {
	MaxWidth int
}

func NewPrettyFormatter() *PrettyFormatter {
	return &PrettyFormatter{MaxWidth: 100}
}

func (f *PrettyFormatter) Format(results []map[string]interface{}) string {
	if len(results) == 0 {
		return "No results"
	}

	var sb strings.Builder

	cols := make([]string, 0)
	for key := range results[0] {
		cols = append(cols, key)
	}

	colWidths := make(map[string]int)
	for _, col := range cols {
		colWidths[col] = len(col)
		for _, row := range results {
			if v, ok := row[col]; ok && v != nil {
				val := fmt.Sprintf("%v", v)
				if len(val) > colWidths[col] {
					colWidths[col] = len(val)
				}
			}
		}
		if colWidths[col] > 30 {
			colWidths[col] = 30
		}
	}

	totalWidth := 0
	for _, w := range colWidths {
		totalWidth += w + 3
	}

	border := "┌" + strings.Repeat("─", totalWidth-2) + "┐"
	bottom := "└" + strings.Repeat("─", totalWidth-2) + "┘"
	divider := "├" + strings.Repeat("─", totalWidth-2) + "┤"

	sb.WriteString(border + "\n")

	sb.WriteString("│")
	for i, col := range cols {
		padding := colWidths[col] - len(col)
		sb.WriteString(" ")
		sb.WriteString(col)
		sb.WriteString(strings.Repeat(" ", padding+1))
		if i < len(cols)-1 {
			sb.WriteString("│")
		}
	}
	sb.WriteString("│\n")
	sb.WriteString(divider + "\n")

	for _, row := range results {
		sb.WriteString("│")
		for i, col := range cols {
			val := ""
			if v, ok := row[col]; ok && v != nil {
				val = fmt.Sprintf("%v", v)
				if len(val) > 30 {
					val = val[:27] + "..."
				}
			}
			padding := colWidths[col] - len(val)
			sb.WriteString(" ")
			sb.WriteString(val)
			sb.WriteString(strings.Repeat(" ", padding+1))
			if i < len(cols)-1 {
				sb.WriteString("│")
			}
		}
		sb.WriteString("│\n")
	}

	sb.WriteString(bottom)

	return sb.String()
}

func GetFormatter(format OutputFormat) Formatter {
	switch format {
	case FormatJSON:
		return NewJSONFormatter(false)
	case FormatPretty:
		return NewPrettyFormatter()
	case FormatCSV:
		return NewCSVFormatter()
	case FormatMarkdown:
		return NewMarkdownFormatter()
	default:
		return nil
	}
}
