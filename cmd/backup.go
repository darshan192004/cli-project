package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"dataset-cli/internal/analyzer"
	"dataset-cli/internal/database"
	"dataset-cli/internal/query"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
)

var backupPath string

var backupCmd = &cobra.Command{
	Use:   "backup [table]",
	Short: "Backup database or table to file",
	Long: `Backup database or specific table to JSON file.

Examples:
  dataset-cli backup                    # Backup all tables
  dataset-cli backup users              # Backup specific table
  dataset-cli backup --output ./backups  # Custom output path`,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := getDB()
		if err != nil {
			PrintError("%v", err)
			os.Exit(1)
		}
		ctx := context.Background()

		if len(args) > 0 {
			backupSingleTable(ctx, db, args[0])
		} else {
			backupAllTables(ctx, db)
		}
	},
}

func backupSingleTable(ctx context.Context, db database.Backend, tableName string) {
	exists, err := db.TableExists(ctx, tableName)
	if err != nil || !exists {
		color.Red.Printf("Table '%s' does not exist\n", tableName)
		return
	}

	exec := query.NewExecutor(db)
	query := fmt.Sprintf("SELECT * FROM \"%s\"", tableName)

	results, err := exec.Execute(query)
	if err != nil {
		PrintError("Backup failed: %v", err)
		os.Exit(1)
	}

	backup := BackupData{
		Version:   "1.0",
		TableName: tableName,
		Exported:  time.Now().Format(time.RFC3339),
		RowCount:  len(results),
		Data:      results,
	}

	data, _ := json.MarshalIndent(backup, "", "  ")

	outputPath := backupPath
	if outputPath == "" {
		outputPath = fmt.Sprintf("%s_%s.json", tableName, time.Now().Format("20060102_150405"))
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		PrintError("Failed to write backup: %v", err)
		return
	}

	color.Green.Printf("Backed up %d rows from table '%s' to %s\n", len(results), tableName, outputPath)
}

func backupAllTables(ctx context.Context, db database.Backend) {
	tables, err := db.GetAllTables(ctx)
	if err != nil {
		PrintError("Failed to get tables: %v", err)
		return
	}

	if len(tables) == 0 {
		color.Yellow.Println("No tables to backup")
		return
	}

	color.Bold.Printf("\nBacking up %d tables...\n\n", len(tables))

	backupDir := backupPath
	if backupDir == "" {
		backupDir = fmt.Sprintf("backup_%s", time.Now().Format("20060102_150405"))
	}

	_ = os.MkdirAll(backupDir, 0755)

	for _, table := range tables {
		color.Cyan.Printf("  Backing up: %s", table)

		exec := query.NewExecutor(db)
		queryStr := fmt.Sprintf("SELECT * FROM \"%s\"", table)

		results, err := exec.Execute(queryStr)
		if err != nil {
			color.Red.Printf(" [FAILED: %v]\n", err)
			continue
		}

		backup := BackupData{
			Version:   "1.0",
			TableName: table,
			Exported:  time.Now().Format(time.RFC3339),
			RowCount:  len(results),
			Data:      results,
		}

		data, _ := json.MarshalIndent(backup, "", "  ")
		outputPath := filepath.Join(backupDir, table+".json")

		if err := os.WriteFile(outputPath, data, 0644); err != nil {
			color.Red.Printf(" [FAILED: %v]\n", err)
			continue
		}

		color.Green.Printf(" [%d rows]\n", len(results))
	}

	color.Green.Printf("\nBackup complete! Files saved to: %s\n", backupDir)
}

type BackupData struct {
	Version   string                   `json:"version"`
	TableName string                   `json:"table_name"`
	Exported  string                   `json:"exported_at"`
	RowCount  int                      `json:"row_count"`
	Data      []map[string]interface{} `json:"data"`
}

var restoreCmd = &cobra.Command{
	Use:   "restore <file>",
	Short: "Restore data from backup file",
	Long: `Restore data from a backup file.

Examples:
  dataset-cli restore backup.json
  dataset-cli restore backup.json --drop  # Drop existing table first`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		backupFile := args[0]

		data, err := os.ReadFile(backupFile)
		if err != nil {
			PrintError("Failed to read backup file: %v", err)
			os.Exit(1)
		}

		var backup BackupData
		if err := json.Unmarshal(data, &backup); err != nil {
			PrintError("Invalid backup file format: %v", err)
			os.Exit(1)
		}

		db, err := getDB()
		if err != nil {
			PrintError("%v", err)
			os.Exit(1)
		}

		ctx := context.Background()
		migrator := database.NewMigrator(db)

		color.Bold.Printf("Restoring backup: %s\n", backup.TableName)
		color.Cyan.Printf("Rows: %d, Exported: %s\n\n", backup.RowCount, backup.Exported)

		if dropTable {
			color.Yellow.Printf("Dropping existing table: %s\n", backup.TableName)
			_, _ = db.Exec(ctx, fmt.Sprintf("DROP TABLE IF EXISTS \"%s\"", backup.TableName))
		}

		schema := &analyzer.Schema{
			TableName: backup.TableName,
			Columns:   inferSchemaFromData(backup.Data),
		}

		if err := migrator.CreateTable(ctx, schema); err != nil {
			PrintError("Failed to create table: %v", err)
			os.Exit(1)
		}

		color.Cyan.Printf("Importing %d rows...\n", len(backup.Data))
		result, err := migrator.ImportData(ctx, schema, backup.Data)
		if err != nil {
			PrintError("Import failed: %v", err)
			os.Exit(1)
		}

		color.Green.Printf("\nRestored %d rows to '%s'\n", result.SuccessCount, backup.TableName)
		if result.ErrorCount > 0 {
			color.Yellow.Printf("Skipped %d rows due to errors\n", result.ErrorCount)
		}
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
	rootCmd.AddCommand(restoreCmd)

	backupCmd.Flags().StringVarP(&backupPath, "output", "o", "", "Output file or directory")
	restoreCmd.Flags().BoolVar(&dropTable, "drop", false, "Drop existing table before restore")
}

func inferSchemaFromData(data []map[string]interface{}) []analyzer.Column {
	if len(data) == 0 {
		return []analyzer.Column{}
	}

	columns := make([]analyzer.Column, 0)
	for key := range data[0] {
		columns = append(columns, analyzer.Column{
			Name: key,
			Type: analyzer.TypeString,
		})
	}
	return columns
}
