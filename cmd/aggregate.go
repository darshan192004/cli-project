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

var (
	groupByColumns []string
	aggFunctions   []string
	aggColumn      string
	sortBy         string
	sortOrder      string
)

var aggregateCmd = &cobra.Command{
	Use:   "aggregate <table>",
	Short: "Aggregate data with count, sum, avg, min, max",
	Long: `Perform aggregations on table data.

Examples:
  dataset-cli aggregate users --count --group-by city
  dataset-cli aggregate sales --sum amount --group-by product
  dataset-cli aggregate orders --avg price --group-by status
  dataset-cli aggregate products --count --sum stock --group-by category`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		tableName := args[0]

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

		schema, _ := db.GetTableSchema(ctx, tableName)

		queryStr := buildAggregateQuery(tableName, schema)

		if DryRun {
			color.Yellow.Printf("Query that would be executed:\n%s\n", queryStr)
			return
		}

		exec := query.NewExecutor(db)
		results, err := exec.Execute(queryStr)
		if err != nil {
			PrintError("Query failed: %v", err)
			os.Exit(1)
		}

		if len(results) == 0 {
			color.Yellow.Println("No results")
			return
		}

		color.Bold.Printf("\n=== Aggregation Results ===\n\n")
		PrintTable(results)
	},
}

func buildAggregateQuery(tableName string, schema []database.ColumnInfo) string {
	var cols []string
	var groupCols []string

	if len(groupByColumns) > 0 {
		for _, gb := range groupByColumns {
			groupCols = append(groupCols, fmt.Sprintf("\"%s\"", gb))
		}
		cols = append(cols, groupCols...)
	}

	for _, fn := range aggFunctions {
		switch strings.ToLower(fn) {
		case "count":
			col := aggColumn
			if col == "" {
				col = "*"
			}
			cols = append(cols, fmt.Sprintf("COUNT(\"%s\") as count", col))
		case "sum":
			if aggColumn != "" {
				cols = append(cols, fmt.Sprintf("SUM(\"%s\") as sum_%s", aggColumn, aggColumn))
			}
		case "avg":
			if aggColumn != "" {
				cols = append(cols, fmt.Sprintf("AVG(\"%s\") as avg_%s", aggColumn, aggColumn))
			}
		case "min":
			if aggColumn != "" {
				cols = append(cols, fmt.Sprintf("MIN(\"%s\") as min_%s", aggColumn, aggColumn))
			}
		case "max":
			if aggColumn != "" {
				cols = append(cols, fmt.Sprintf("MAX(\"%s\") as max_%s", aggColumn, aggColumn))
			}
		}
	}

	if len(cols) == 0 {
		cols = append(cols, "COUNT(*)")
	}

	queryStr := fmt.Sprintf("SELECT %s FROM \"%s\"", strings.Join(cols, ", "), tableName)

	if len(groupCols) > 0 {
		queryStr += fmt.Sprintf(" GROUP BY %s", strings.Join(groupCols, ", "))
	}

	if sortBy != "" {
		order := "ASC"
		if strings.ToLower(sortOrder) == "desc" {
			order = "DESC"
		}
		queryStr += fmt.Sprintf(" ORDER BY \"%s\" %s", sortBy, order)
	}

	return queryStr
}

func init() {
	rootCmd.AddCommand(aggregateCmd)

	aggregateCmd.Flags().StringSliceVar(&groupByColumns, "group-by", []string{}, "Columns to group by")
	aggregateCmd.Flags().StringSliceVar(&aggFunctions, "count", []string{}, "Aggregation functions (count, sum, avg, min, max)")
	aggregateCmd.Flags().StringVar(&aggColumn, "column", "", "Column to aggregate")
	aggregateCmd.Flags().StringVar(&sortBy, "sort-by", "", "Column to sort by")
	aggregateCmd.Flags().StringVar(&sortOrder, "order", "asc", "Sort order (asc/desc)")
}
