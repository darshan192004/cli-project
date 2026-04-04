package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"dataset-cli/internal/analyzer"
	"dataset-cli/internal/database"
	"dataset-cli/internal/reader"
	"github.com/spf13/cobra"
)

var (
	filePath     string
	tableName    string
	dropTable    bool
	skipErrors   bool
	showProgress bool
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
  dataset-cli migrate data.csv --drop
  dataset-cli migrate data.csv --skip-errors (continue on data errors)`,
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
			fmt.Printf("  - %s: %s\n", col.Name, col.Type)
		}

		if dropTable {
			fmt.Printf("Dropping existing table: %s\n", schema.TableName)
			_, err := db.Exec(context.Background(), fmt.Sprintf("DROP TABLE IF EXISTS %s", schema.TableName))
			if err != nil {
				fmt.Printf("Error dropping table: %v\n", err)
				os.Exit(1)
			}
		}

		migratorOpts := []database.MigratorOption{
			database.WithSkipErrors(skipErrors),
			database.WithBatchSize(1000),
		}

		if showProgress {
			migratorOpts = append(migratorOpts, database.WithProgressCallback(func(current, total int) {
				percent := float64(current) / float64(total) * 100
				fmt.Printf("\rProgress: %d/%d (%.1f%%)", current, total, percent)
			}))
		}

		migrator := database.NewMigrator(db, migratorOpts...)
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
		startTime := time.Now()

		if showProgress {
			fmt.Println()
		}

		result, err := migrator.ImportData(context.Background(), schema, records)
		if err != nil {
			fmt.Printf("\nError importing data: %v\n", err)
			os.Exit(1)
		}

		elapsed := time.Since(startTime)

		if showProgress {
			fmt.Println()
		}

		if skipErrors && result.ErrorCount > 0 {
			fmt.Printf("\nImport completed with errors:\n")
			fmt.Printf("  - Successful: %d\n", result.SuccessCount)
			fmt.Printf("  - Failed: %d\n", result.ErrorCount)
			fmt.Printf("  - Time: %v\n", elapsed.Round(time.Second))

			fmt.Println("\nFirst 5 errors:")
			for i := 0; i < 5 && i < len(result.Errors); i++ {
				err := result.Errors[i]
				fmt.Printf("  Record %d: %s\n", err.RecordIndex+1, err.Error)
			}
		} else {
			fmt.Printf("\nSuccessfully imported %d records into table: %s\n", result.SuccessCount, schema.TableName)
			fmt.Printf("Time: %v\n", elapsed.Round(time.Second))
		}

		fmt.Printf("\nUse the following commands to query your data:\n")
		fmt.Printf("  dataset-cli filter %s --where \"condition\"\n", schema.TableName)
		fmt.Printf("  dataset-cli paginate %s\n", schema.TableName)
		fmt.Printf("  dataset-cli schema %s\n", schema.TableName)
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)

	migrateCmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to CSV or JSON file")
	migrateCmd.Flags().StringVar(&tableName, "table-name", "", "Custom table name")
	migrateCmd.Flags().BoolVar(&dropTable, "drop", false, "Drop existing table before import")
	migrateCmd.Flags().BoolVar(&skipErrors, "skip-errors", false, "Skip rows with errors instead of failing")
	migrateCmd.Flags().BoolVar(&showProgress, "progress", true, "Show import progress")
}
