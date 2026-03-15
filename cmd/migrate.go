package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"dataset-cli/internal/analyzer"
	"dataset-cli/internal/database"
	"dataset-cli/internal/reader"
)

var (
	filePath string
	tableName string
	dropTable bool
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Import CSV/JSON files into PostgreSQL",
	Long: `Analyze a CSV or JSON file and import it into PostgreSQL.

The command will:
1. Read the file and detect column types
2. Create a table with appropriate PostgreSQL types
3. Import all data into the table

Example:
  dataset-cli migrate data.csv
  dataset-cli migrate data.json --table-name users
  dataset-cli migrate data.csv --drop`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 && filePath == "" {
			fmt.Println("Error: please provide a file path")
			os.Exit(1)
		}

		if filePath == "" {
			filePath = args[0]
		}

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			fmt.Printf("Error: file not found: %s\n", filePath)
			os.Exit(1)
		}

		db, err := getDB()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		fmt.Printf("Analyzing file: %s\n", filePath)

		analyzer := analyzer.New()
		schema, err := analyzer.Analyze(filePath)
		if err != nil {
			fmt.Printf("Error analyzing file: %v\n", err)
			os.Exit(1)
		}

		if tableName != "" {
			schema.TableName = tableName
		}

		fmt.Printf("Detected schema for table: %s\n", schema.TableName)
		fmt.Println("Columns:")
		for _, col := range schema.Columns {
			pk := ""
			if col.IsPrimaryKey {
				pk = " (PRIMARY KEY)"
			}
			fmt.Printf("  - %s: %s%s\n", col.Name, col.Type, pk)
		}

		if dropTable {
			fmt.Printf("Dropping existing table: %s\n", schema.TableName)
			_, err := db.Pool.Exec(context.Background(), fmt.Sprintf("DROP TABLE IF EXISTS %s", schema.TableName))
			if err != nil {
				fmt.Printf("Error dropping table: %v\n", err)
				os.Exit(1)
			}
		}

		migrator := database.NewMigrator(db)
		if err := migrator.CreateTable(context.Background(), schema); err != nil {
			fmt.Printf("Error creating table: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Reading data from file...")
		r, err := reader.GetReader(filePath)
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			os.Exit(1)
		}

		records, err := r.Read(filePath)
		if err != nil {
			fmt.Printf("Error parsing file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Importing %d records...\n", len(records))

		if err := migrator.ImportData(context.Background(), schema, records); err != nil {
			fmt.Printf("Error importing data: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully imported %d records into table: %s\n", len(records), schema.TableName)
		fmt.Printf("\nUse the following commands to query your data:\n")
		fmt.Printf("  dataset-cli filter %s --where \"condition\"\n", schema.TableName)
		fmt.Printf("  dataset-cli paginate %s\n", schema.TableName)
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)

	migrateCmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to CSV or JSON file")
	migrateCmd.Flags().StringVar(&tableName, "table-name", "", "Custom table name")
	migrateCmd.Flags().BoolVar(&dropTable, "drop", false, "Drop existing table before import")
}
