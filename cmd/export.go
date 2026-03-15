package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"dataset-cli/internal/query"
)

var (
	outputPath string
	outputFormat string
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export query results to JSON or CSV",
	Long: `Export data from a table to a file.

Example:
  dataset-cli export users --output data.json
  dataset-cli export users --output data.csv --format csv
  dataset-cli export products --output filtered.json --where "price > 100"`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Error: please provide a table name")
			os.Exit(1)
		}

		tableName := args[0]

		if outputPath == "" {
			fmt.Println("Error: please provide an output path with --output")
			os.Exit(1)
		}

		db, err := getDB()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		exists, err := db.TableExists(nil, tableName)
		if err != nil || !exists {
			fmt.Printf("Error: table '%s' does not exist\n", tableName)
			os.Exit(1)
		}

		q := query.New(tableName)
		exec := query.NewExecutor(db)

		queryStr, _ := q.Filter(whereClause)

		results, err := exec.Execute(queryStr)
		if err != nil {
			fmt.Printf("Error executing query: %v\n", err)
			os.Exit(1)
		}

		if len(results) == 0 {
			fmt.Println("No results to export")
			return
		}

		ext := strings.ToLower(filepath.Ext(outputPath))
		format := outputFormat
		if format == "" {
			switch ext {
			case ".json":
				format = "json"
			case ".csv":
				format = "csv"
			default:
				format = "json"
			}
		}

		var writeErr error
		switch format {
		case "json":
			writeErr = writeJSON(outputPath, results)
		case "csv":
			writeErr = writeCSV(outputPath, results)
		default:
			fmt.Printf("Error: unsupported format '%s' (supported: json, csv)\n", format)
			os.Exit(1)
		}

		if writeErr != nil {
			fmt.Printf("Error writing file: %v\n", writeErr)
			os.Exit(1)
		}

		fmt.Printf("Successfully exported %d records to %s\n", len(results), outputPath)
	},
}

func writeJSON(path string, results []map[string]interface{}) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func writeCSV(path string, results []map[string]interface{}) error {
	if len(results) == 0 {
		return nil
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	var headers []string
	for key := range results[0] {
		headers = append(headers, key)
	}
	writer.Write(headers)

	for _, row := range results {
		var record []string
		for _, h := range headers {
			val := ""
			if v, ok := row[h]; ok {
				val = fmt.Sprintf("%v", v)
			}
			record = append(record, val)
		}
		writer.Write(record)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output file path")
	exportCmd.Flags().StringVar(&outputFormat, "format", "", "Output format (json, csv)")
	exportCmd.Flags().StringVar(&whereClause, "where", "", "WHERE condition")
}
