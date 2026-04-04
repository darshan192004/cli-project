package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Show table schema and column information",
	Long: `Display the schema of a table including columns, types, and row count.

Example:
  dataset-cli schema users
  dataset-cli schema products`,
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
		info, err := db.GetTableInfo(ctx, tableName)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Table: %s\n", info.Name)
		fmt.Printf("Total Rows: %d\n\n", info.Count)
		fmt.Println("Columns:")
		fmt.Println("+----------------------+------------------+------------+-----------------+")
		fmt.Println("| Column Name          | Data Type        | Nullable   | Default         |")
		fmt.Println("+----------------------+------------------+------------+-----------------+")

		for _, col := range info.Columns {
			nullable := col.IsNullable
			defaultVal := ""
			if col.Default != nil {
				defaultVal = *col.Default
			}
			fmt.Printf("| %-20s | %-16s | %-10s | %-15s |\n",
				col.Name, col.DataType, nullable, defaultVal)
		}
		fmt.Println("+----------------------+------------------+------------+-----------------+")
	},
}

func init() {
	rootCmd.AddCommand(schemaCmd)
}
