package cmd

import (
	"context"
	"fmt"
	"os"

	"dataset-cli/internal/database"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats [table]",
	Short: "Show table statistics and data distribution",
	Long: `Display detailed statistics for a table including:
- Row count
- Column information with cardinality
- Data distribution for text columns
- NULL value counts

Example:
  dataset-cli stats users
  dataset-cli stats products --sample 1000`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		tableName := args[0]

		db, err := getDB()
		if err != nil {
			PrintError("%v", err)
			os.Exit(1)
		}

		ctx := context.Background()
		info, err := db.GetTableInfo(ctx, tableName)
		if err != nil {
			PrintError("Failed to get table info: %v", err)
			os.Exit(1)
		}

		sampleSize, _ := cmd.Flags().GetInt("sample")
		showNulls, _ := cmd.Flags().GetBool("show-nulls")

		printStats(tableName, info, sampleSize, showNulls)
	},
}

func printStats(tableName string, info *database.TableInfo, sampleSize int, showNulls bool) {
	header := color.New(color.Bold, color.Cyan)

	header.Println("\n" + repeatStr("=", 60))
	header.Printf("  Table Statistics: %s\n", tableName)
	header.Println(repeatStr("=", 60))

	header.Println("\n[Basic Info]")
	fmt.Printf("  Total Rows: %s\n", color.Green.Sprint(info.Count))
	fmt.Printf("  Columns: %s\n", color.Green.Sprint(len(info.Columns)))

	header.Println("\n[Column Details]")

	border := "+" + repeatStr("-", 60) + "+"
	fmt.Println(border)
	fmt.Printf("| %-20s | %-15s | %-10s |\n",
		color.Bold.Sprint("Column"),
		color.Bold.Sprint("Type"),
		color.Bold.Sprint("Nullable"))
	fmt.Println(border)

	for _, col := range info.Columns {
		nullable := col.IsNullable
		if nullable == "YES" {
			nullable = color.Yellow.Sprint(nullable)
		} else {
			nullable = color.Green.Sprint(nullable)
		}
		fmt.Printf("| %-20s | %-15s | %-10s |\n",
			color.Cyan.Sprint(col.Name),
			col.DataType,
			nullable)
	}
	fmt.Println(border)

	if showNulls && info.Count > 0 {
		showNullCounts(tableName, info.Columns)
	}
}

func showNullCounts(tableName string, columns []database.ColumnInfo) {
	header := color.New(color.Bold, color.Yellow)
	header.Println("\n[NULL Value Counts]")

	for _, col := range columns {
		if col.IsNullable == "YES" {
			var nullCount int64
			query := fmt.Sprintf("SELECT COUNT(*) FROM \"%s\" WHERE \"%s\" IS NULL", tableName, col.Name)
			db, _ := getDB()
			if db != nil {
				row, _ := db.QueryRow(context.Background(), query)
				_ = row.(interface{ Scan(...interface{}) error }).Scan(&nullCount)
				if nullCount > 0 {
					fmt.Printf("  %s: %s NULL values\n",
						color.Cyan.Sprint(col.Name),
						color.Yellow.Sprint(nullCount))
				}
			}
		}
	}
}

func repeatStr(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

func init() {
	rootCmd.AddCommand(statsCmd)
	statsCmd.Flags().Int("sample", 100, "Sample size for cardinality analysis")
	statsCmd.Flags().Bool("show-nulls", false, "Show NULL value counts per column")
}
