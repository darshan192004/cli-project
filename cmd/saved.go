package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"dataset-cli/internal/query"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
)

type SavedQuery struct {
	Name        string `json:"name"`
	Table       string `json:"table"`
	Columns     string `json:"columns,omitempty"`
	Where       string `json:"where,omitempty"`
	Limit       int    `json:"limit,omitempty"`
	Description string `json:"description,omitempty"`
}

var (
	savedTableName   string
	savedWhereClause string
	savedColumnsList string
	savedQueryLimit  int
	queryDescription string
)

var savedCmd = &cobra.Command{
	Use:   "saved",
	Short: "Manage saved queries",
	Long: `Manage and run saved queries.

Examples:
  dataset-cli saved list
  dataset-cli saved add my-query --table users --where "age > 25"
  dataset-cli saved run my-query
  dataset-cli saved delete my-query`,
}

var savedListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all saved queries",
	Run: func(cmd *cobra.Command, args []string) {
		queries := loadSavedQueries()

		if len(queries) == 0 {
			color.Yellow.Println("No saved queries")
			return
		}

		color.Bold.Println("\n=== Saved Queries ===\n")

		for i, q := range queries {
			color.Cyan.Printf("  %d. %s\n", i+1, q.Name)
			fmt.Printf("     Table: %s\n", q.Table)
			if q.Where != "" {
				fmt.Printf("     Filter: %s\n", q.Where)
			}
			if q.Description != "" {
				color.Gray.Printf("     %s\n", q.Description)
			}
			fmt.Println()
		}
	},
}

var savedAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Save a query",
	Long:  `Save a query for quick access.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		queries := loadSavedQueries()

		for _, q := range queries {
			if q.Name == name {
				color.Red.Printf("Query '%s' already exists\n", name)
				return
			}
		}

		savedQuery := SavedQuery{
			Name:        name,
			Table:       savedTableName,
			Columns:     savedColumnsList,
			Where:       savedWhereClause,
			Limit:       savedQueryLimit,
			Description: queryDescription,
		}

		queries = append(queries, savedQuery)
		saveSavedQueries(queries)

		color.Green.Printf("Query '%s' saved successfully\n", name)
	},
}

var savedRunCmd = &cobra.Command{
	Use:   "run <name>",
	Short: "Run a saved query",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		queries := loadSavedQueries()

		var selectedQuery *SavedQuery
		for i := range queries {
			if queries[i].Name == name {
				selectedQuery = &queries[i]
				break
			}
		}

		if selectedQuery == nil {
			color.Red.Printf("Query '%s' not found\n", name)
			return
		}

		db, err := getDB()
		if err != nil {
			PrintError("%v", err)
			os.Exit(1)
		}

		queryStr := buildSavedQuery(selectedQuery)
		exec := query.NewExecutor(db)

		if DryRun {
			color.Yellow.Printf("Query that would be executed:\n%s\n", queryStr)
			return
		}

		results, err := exec.Execute(queryStr)
		if err != nil {
			PrintError("Query failed: %v", err)
			os.Exit(1)
		}

		if len(results) == 0 {
			color.Yellow.Println("No results")
			return
		}

		color.Green.Printf("Query '%s' returned %d results\n\n", name, len(results))
		PrintTable(results)
	},
}

var savedDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a saved query",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		queries := loadSavedQueries()

		newQueries := make([]SavedQuery, 0)
		found := false
		for _, q := range queries {
			if q.Name == name {
				found = true
			} else {
				newQueries = append(newQueries, q)
			}
		}

		if !found {
			color.Red.Printf("Query '%s' not found\n", name)
			return
		}

		saveSavedQueries(newQueries)
		color.Green.Printf("Query '%s' deleted\n", name)
	},
}

func getSavedQueriesPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".dataset-cli", "saved-queries.json")
}

func loadSavedQueries() []SavedQuery {
	path := getSavedQueriesPath()

	data, err := os.ReadFile(path)
	if err != nil {
		return []SavedQuery{}
	}

	var queries []SavedQuery
	if err := json.Unmarshal(data, &queries); err != nil {
		return []SavedQuery{}
	}

	return queries
}

func saveSavedQueries(queries []SavedQuery) {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".dataset-cli")
	os.MkdirAll(dir, 0755)

	path := getSavedQueriesPath()
	data, _ := json.MarshalIndent(queries, "", "  ")
	os.WriteFile(path, data, 0644)
}

func buildSavedQuery(q *SavedQuery) string {
	cols := "*"
	if q.Columns != "" {
		cols = q.Columns
	}

	queryStr := fmt.Sprintf("SELECT %s FROM \"%s\"", cols, q.Table)
	if q.Where != "" {
		queryStr += " WHERE " + q.Where
	}
	if q.Limit > 0 {
		queryStr += fmt.Sprintf(" LIMIT %d", q.Limit)
	}

	return queryStr
}

func init() {
	rootCmd.AddCommand(savedCmd)
	savedCmd.AddCommand(savedListCmd)
	savedCmd.AddCommand(savedAddCmd)
	savedCmd.AddCommand(savedRunCmd)
	savedCmd.AddCommand(savedDeleteCmd)

	savedAddCmd.Flags().StringVar(&savedTableName, "table", "", "Table name (required)")
	savedAddCmd.Flags().StringVar(&savedWhereClause, "where", "", "WHERE condition")
	savedAddCmd.Flags().StringVar(&savedColumnsList, "columns", "", "Columns to select")
	savedAddCmd.Flags().IntVar(&savedQueryLimit, "limit", 0, "Result limit")
	savedAddCmd.Flags().StringVar(&queryDescription, "description", "", "Description")
	savedAddCmd.MarkFlagRequired("table")
}
