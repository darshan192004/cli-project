package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"dataset-cli/internal/database"
	"dataset-cli/internal/query"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
)

var searchTerm string

var searchCmd = &cobra.Command{
	Use:   "search <table> <term>",
	Short: "Search across all text columns in a table",
	Long: `Search for a term across all text columns in a table.

This is useful when you don't know which column contains the data you're looking for.

Examples:
  dataset-cli search users "john"
  dataset-cli search products "organic"
  dataset-cli search orders "pending" --limit 50`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		tableName := args[0]
		searchTerm = args[1]

		db, err := getDB()
		if err != nil {
			PrintError("%v", err)
			os.Exit(1)
		}

		ctx := context.Background()
		exists, err := db.TableExists(ctx, tableName)
		if err != nil || !exists {
			PrintError("Table '%s' does not exist", tableName)
			os.Exit(1)
		}

		schema, err := db.GetTableSchema(ctx, tableName)
		if err != nil {
			PrintError("Failed to get table schema: %v", err)
			os.Exit(1)
		}

		textCols := getTextColumns(schema)

		if len(textCols) == 0 {
			color.Yellow.Println("No text columns found in table")
			return
		}

		queryStr := buildSearchQuery(tableName, textCols)

		if DryRun {
			color.Yellow.Printf("Query that would be executed:\n%s\n", queryStr)
			return
		}

		exec := query.NewExecutor(db)
		results, err := exec.Execute(queryStr)
		if err != nil {
			PrintError("Search failed: %v", err)
			os.Exit(1)
		}

		if len(results) == 0 {
			color.Yellow.Printf("No results found for '%s'\n", searchTerm)
			return
		}

		color.Green.Printf("Found %d results for '%s'\n\n", len(results), searchTerm)
		PrintTable(results)
	},
}

func getTextColumns(schema []database.ColumnInfo) []string {
	var textCols []string
	for _, col := range schema {
		if strings.Contains(strings.ToLower(col.DataType), "text") ||
			strings.Contains(strings.ToLower(col.DataType), "varchar") ||
			strings.Contains(strings.ToLower(col.DataType), "char") {
			textCols = append(textCols, col.Name)
		}
	}
	return textCols
}

func buildSearchQuery(tableName string, textCols []string) string {
	var conditions []string
	for _, col := range textCols {
		conditions = append(conditions, fmt.Sprintf("CAST(\"%s\" AS TEXT) ILIKE '%%%s%%'", col, searchTerm))
	}

	query := fmt.Sprintf("SELECT * FROM \"%s\" WHERE %s", tableName, strings.Join(conditions, " OR "))

	if limitCount > 0 {
		query += fmt.Sprintf(" LIMIT %d", limitCount)
	}

	return query
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().IntVar(&limitCount, "limit", 100, "Maximum number of results")
}
