package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"dataset-cli/internal/query"
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
		defer db.Close()

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
		if err := db.Pool.QueryRow(ctx, countQuery).Scan(&totalCount); err != nil {
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
			fmt.Println("No results found")
			return
		}

		remaining := int(totalCount) - (page * limit)
		if remaining < 0 {
			remaining = 0
		}

		fmt.Printf("\n=== Filter Results ===\n")
		fmt.Printf("Page %d of %d (Total: %d records, %d remaining)\n\n", page, totalPages, totalCount, remaining)
		printResults(results)

		if totalPages > page {
			fmt.Printf("\nNext page: dataset-cli filter %s --where \"%s\" --limit %d --page %d\n",
				tableName, whereClause, limit, page+1)
		}
	},
}

func printResults(results []map[string]interface{}) {
	if len(results) == 0 {
		return
	}

	cols := make([]string, 0)
	for key := range results[0] {
		cols = append(cols, key)
	}

	colWidths := make(map[string]int)
	for _, col := range cols {
		l := len(col)
		for _, row := range results {
			if v, ok := row[col]; ok {
				vStr := fmt.Sprintf("%v", v)
				if len(vStr) > l {
					l = len(vStr)
				}
			}
		}
		if l > 30 {
			l = 30
		}
		colWidths[col] = l + 2
	}

	totalWidth := 1
	for _, w := range colWidths {
		totalWidth += w + 1
	}

	fmt.Println("+" + strings.Repeat("-", totalWidth-1) + "+")
	fmt.Print("|")
	for _, col := range cols {
		displayCol := col
		if len(col) > colWidths[col]-2 {
			displayCol = col[:colWidths[col]-5] + "..."
		}
		fmt.Printf(" %-*s |", colWidths[col]-1, displayCol)
	}
	fmt.Println()
	fmt.Println("+" + strings.Repeat("-", totalWidth-1) + "+")

	for _, row := range results {
		fmt.Print("|")
		for _, col := range cols {
			val := ""
			if v, ok := row[col]; ok {
				val = fmt.Sprintf("%v", v)
				if len(val) > colWidths[col]-2 {
					val = val[:colWidths[col]-5] + "..."
				}
			}
			fmt.Printf(" %-*s |", colWidths[col]-1, val)
		}
		fmt.Println()
	}
	fmt.Println("+" + strings.Repeat("-", totalWidth-1) + "+")
}

func askExport(results []map[string]interface{}) {
	fmt.Print("\nDo you want to export results to a file? (y/n): ")
	var response string
	fmt.Scanln(&response)

	if response == "y" || response == "Y" {
		fmt.Print("Enter output file path (e.g., output.json): ")
		var outputPath string
		fmt.Scanln(&outputPath)

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
