package cmd

import (
	"context"
	"fmt"
	"os"

	"dataset-cli/internal/query"
	"github.com/spf13/cobra"
)

var (
	page     int
	pageSize int
)

func paginateWithContext(ctx context.Context, tableName string, pageNum, size int) {
	if pageNum < 1 {
		pageNum = 1
	}
	if size < 1 {
		size = 10
	}

	db, err := getDB()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	exists, err := db.TableExists(ctx, tableName)
	if err != nil || !exists {
		fmt.Printf("Error: table '%s' does not exist\n", tableName)
		return
	}

	offset := (pageNum - 1) * size

	q := query.New(tableName)
	exec := query.NewExecutor(db)

	queryStr, _ := q.Paginate(size, offset, nil)

	results, err := exec.Execute(queryStr)
	if err != nil {
		fmt.Printf("Error executing query: %v\n", err)
		return
	}

	var total int64
	countQuery, _ := q.Count("")
	row, err := db.QueryRow(ctx, countQuery)
	if err == nil && row != nil {
		_ = row.(interface{ Scan(...interface{}) error }).Scan(&total)
	}

	totalPages := int(total) / size
	if int(total)%size > 0 {
		totalPages++
	}

	if len(results) == 0 {
		fmt.Println("No results found")
		return
	}

	fmt.Printf("\nPage %d of %d (Total: %d records)\n\n", pageNum, totalPages, total)
	PrintTable(results)

	if pageNum < totalPages {
		fmt.Printf("\nNext page: dataset-cli paginate %s --page %d --page-size %d\n",
			tableName, pageNum+1, size)
	}
	if pageNum > 1 {
		fmt.Printf("Previous page: dataset-cli paginate %s --page %d --page-size %d\n",
			tableName, pageNum-1, size)
	}
}

var paginateCmd = &cobra.Command{
	Use:   "paginate",
	Short: "Paginated queries with limit and offset",
	Long: `Get paginated results from a table.

Example:
  dataset-cli paginate users
  dataset-cli paginate users --page 2 --page-size 20
  dataset-cli paginate products --columns name,price --page 1 --page-size 50`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Error: please provide a table name")
			os.Exit(1)
		}

		tableName := args[0]

		db, err := getDB()
		if err != nil {
			PrintError("%v", err)
			os.Exit(1)
		}

		ctx := context.Background()
		exists, err := db.TableExists(ctx, tableName)
		if err != nil || !exists {
			fmt.Printf("Error: table '%s' does not exist\n", tableName)
			os.Exit(1)
		}

		if page < 1 {
			page = 1
		}
		if pageSize < 1 {
			pageSize = 10
		}

		offset := (page - 1) * pageSize

		q := query.New(tableName)
		exec := query.NewExecutor(db)

		queryStr, _ := q.Paginate(pageSize, offset, selectColumns)

		results, err := exec.Execute(queryStr)
		if err != nil {
			fmt.Printf("Error executing query: %v\n", err)
			os.Exit(1)
		}

		var total int64
		countQuery, _ := q.Count("")
		row, err := db.QueryRow(ctx, countQuery)
		if err == nil && row != nil {
			_ = row.(interface{ Scan(...interface{}) error }).Scan(&total)
		}

		totalPages := int(total) / pageSize
		if int(total)%pageSize > 0 {
			totalPages++
		}

		if len(results) == 0 {
			fmt.Println("No results found")
			return
		}

		fmt.Printf("Page %d of %d (Total: %d records)\n\n", page, totalPages, total)
		PrintTable(results)

		if page < totalPages {
			fmt.Printf("\nNext page: dataset-cli paginate %s --page %d --page-size %d\n",
				tableName, page+1, pageSize)
		}
		if page > 1 {
			fmt.Printf("Previous page: dataset-cli paginate %s --page %d --page-size %d\n",
				tableName, page-1, pageSize)
		}
	},
}

func init() {
	rootCmd.AddCommand(paginateCmd)

	paginateCmd.Flags().IntVar(&page, "page", 1, "Page number (starting from 1)")
	paginateCmd.Flags().IntVar(&pageSize, "page-size", 10, "Number of records per page")
	paginateCmd.Flags().StringSliceVarP(&selectColumns, "columns", "c", []string{}, "Columns to select")
}
