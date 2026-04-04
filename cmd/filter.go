package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"dataset-cli/internal/query"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
)

var (
	whereClause string
	limitCount  int
	pageNum     int
)

var filterCmd = &cobra.Command{
	Use:   "filter",
	Short: "Filter data with WHERE conditions",
	Long: `Filter data from a table using WHERE conditions.

Example:
  dataset-cli filter users --where "age > 25"
  dataset-cli filter users --where "city = 'New York'" --limit 10
  dataset-cli filter products --where "price < 100"`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Error: please provide a table name")
			os.Exit(1)
		}

		tableName := args[0]

		db, err := getDB()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		ctx := context.Background()
		exists, err := db.TableExists(ctx, tableName)
		if err != nil || !exists {
			fmt.Printf("Error: table '%s' does not exist\n", tableName)
			os.Exit(1)
		}

		exec := query.NewExecutor(db)

		limit := limitCount
		if limit <= 0 {
			limit = 100
		}

		page := pageNum
		if page < 1 {
			page = 1
		}

		offset := (page - 1) * limit

		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM \"%s\"", tableName)
		if whereClause != "" {
			countQuery += " WHERE " + whereClause
		}

		var totalCount int64
		row, _ := db.QueryRow(ctx, countQuery)
		if err := row.(interface{ Scan(...interface{}) error }).Scan(&totalCount); err != nil {
			fmt.Printf("Error counting records: %v\n", err)
			os.Exit(1)
		}

		totalPages := int(totalCount) / limit
		if int(totalCount)%limit > 0 {
			totalPages++
		}

		queryStr := fmt.Sprintf("SELECT * FROM \"%s\"", tableName)
		if whereClause != "" {
			queryStr += " WHERE " + whereClause
		}
		queryStr += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)

		results, err := exec.Execute(queryStr)
		if err != nil {
			fmt.Printf("Error executing query: %v\n", err)
			os.Exit(1)
		}

		if len(results) == 0 {
			color.Yellow.Println("No results found")
			return
		}

		remaining := int(totalCount) - (page * limit)
		if remaining < 0 {
			remaining = 0
		}

		color.Bold.Printf("\n=== Filter Results ===\n")
		fmt.Printf("Page %d of %d (Total: %d records, %d remaining)\n\n", page, totalPages, totalCount, remaining)
		PrintTable(results)

		if totalPages > page {
			fmt.Printf("\nNext page: dataset-cli filter %s --where \"%s\" --limit %d --page %d\n",
				tableName, whereClause, limit, page+1)
		}
	},
}

func askExport(results []map[string]interface{}) {
	fmt.Print("\nDo you want to export results to a file? (y/n): ")
	var response string
	_, _ = fmt.Scanln(&response)

	if response == "y" || response == "Y" {
		fmt.Print("Enter output file path (e.g., output.json): ")
		var outputPath string
		_, _ = fmt.Scanln(&outputPath)

		var data []byte
		if len(results) > 0 {
			data, _ = json.MarshalIndent(results, "", "  ")
		} else {
			data = []byte("[]")
		}

		if err := os.WriteFile(outputPath, data, 0644); err != nil {
			fmt.Printf("Error writing file: %v\n", err)
		} else {
			fmt.Printf("Results exported to %s\n", outputPath)
		}
	}
}

func init() {
	rootCmd.AddCommand(filterCmd)

	filterCmd.Flags().StringVar(&whereClause, "where", "", "WHERE condition (e.g., 'age > 25')")
	filterCmd.Flags().IntVar(&limitCount, "limit", 100, "Maximum number of results per page")
	filterCmd.Flags().IntVar(&pageNum, "page", 1, "Page number")
}
