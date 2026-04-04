package cmd

import (
	"context"
	"fmt"
	"os"

	"dataset-cli/internal/query"
	"github.com/spf13/cobra"
)

var transformCmd = &cobra.Command{
	Use:   "transform",
	Short: "Select specific columns from a table",
	Long: `Select specific columns from a table with optional filtering.

Example:
  dataset-cli transform users --columns name,email
  dataset-cli transform users --columns id,name --where "active = true"`,
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

		q := query.New(tableName)
		exec := query.NewExecutor(db)

		columns := selectColumns
		if len(columns) == 0 {
			schema, _ := db.GetTableSchema(ctx, tableName)
			for _, col := range schema {
				columns = append(columns, col.Name)
			}
		}

		queryStr, _ := q.Transform(columns, whereClause)

		results, err := exec.Execute(queryStr)
		if err != nil {
			fmt.Printf("Error executing query: %v\n", err)
			os.Exit(1)
		}

		if len(results) == 0 {
			fmt.Println("No results found")
			return
		}

		fmt.Printf("Results (%d columns):\n\n", len(columns))
		PrintTable(results)

		askExport(results)
	},
}

func init() {
	rootCmd.AddCommand(transformCmd)

	transformCmd.Flags().StringSliceVarP(&selectColumns, "columns", "c", []string{}, "Comma-separated columns to select (e.g., 'name,email,age')")
	transformCmd.Flags().StringVar(&whereClause, "where", "", "WHERE condition")
}
