package cmd

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"dataset-cli/internal/query"
	"github.com/AlecAivazis/survey/v2"
	"github.com/gookit/color"
)

func StartInteractive() {
	runWizard()
}

func runWizard() {
	for {
		printBanner()

		var operation string
		qs := &survey.Select{
			Message: "What would you like to do?",
			Options: []string{
				"Migrate data (CSV/JSON -> PostgreSQL)",
				"Filter data",
				"Transform/Select columns",
				"Paginate data",
				"View table schema",
				"Export data",
				"Delete table",
				"Exit",
			},
			Default: "Migrate data (CSV/JSON -> PostgreSQL)",
		}
		survey.AskOne(qs, &operation)

		switch operation {
		case "Migrate data (CSV/JSON -> PostgreSQL)":
			interactiveMigrate()
		case "Filter data":
			interactiveFilter()
		case "Transform/Select columns":
			interactiveTransform()
		case "Paginate data":
			interactivePaginate()
		case "View table schema":
			interactiveSchema()
		case "Export data":
			interactiveExport()
		case "Delete table":
			interactiveDelete()
		case "Exit":
			color.Green.Println("\nGoodbye!")
			return
		}

		color.Gray.Println("\nPress Enter to continue...")
		fmt.Scanln()
	}
}

func printBanner() {
	color.Bold.Printf("\n")
	color.Bold.Printf("  %s\n", strings.Repeat("=", 45))
	color.Green.Printf("       Dataset CLI - Interactive Mode")
	color.Bold.Printf("  %s\n", strings.Repeat("=", 45))
	color.Bold.Printf("\n")
}

func getExistingTables() ([]string, error) {
	db, err := getDB()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	tables, err := db.GetAllTables(ctx)
	if err != nil {
		return nil, err
	}
	return tables, nil
}

func getTableColumns(tableName string) ([]string, error) {
	db, err := getDB()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	info, err := db.GetTableInfo(ctx, tableName)
	if err != nil {
		return nil, err
	}

	columns := make([]string, len(info.Columns))
	for i, col := range info.Columns {
		columns[i] = col.Name
	}
	return columns, nil
}

func interactiveMigrate() {
	qs := []*survey.Question{
		{
			Name: "filePath",
			Prompt: &survey.Input{
				Message: "Enter path to CSV or JSON file:",
				Help:    "The file must exist and be a valid CSV or JSON file",
			},
			Validate: survey.Required,
		},
		{
			Name: "tableName",
			Prompt: &survey.Input{
				Message: "Enter table name (or press Enter to use filename):",
			},
		},
		{
			Name: "dropTable",
			Prompt: &survey.Confirm{
				Message: "Drop existing table if it exists?",
				Default: false,
			},
		},
		{
			Name: "skipErrors",
			Prompt: &survey.Confirm{
				Message: "Skip rows with errors? (Recommended for messy data)",
				Default: true,
			},
		},
	}

	answers := struct {
		FilePath   string
		TableName  string
		DropTable  bool
		SkipErrors bool
	}{}

	err := survey.Ask(qs, &answers)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	dropTable = answers.DropTable
	skipErrors = answers.SkipErrors

	fmt.Println("\nRunning migration...")

	args := []string{answers.FilePath}
	if answers.TableName != "" {
		tableName = answers.TableName
	}

	migrateCmd.Run(nil, args)
}

func interactiveFilter() {
	tables, err := getExistingTables()
	if err != nil {
		fmt.Println("Error getting tables:", err)
		return
	}

	if len(tables) == 0 {
		fmt.Println("No tables found in database. Please migrate a file first.")
		return
	}

	var tableName string
	qs := &survey.Select{
		Message: "Select table to filter:",
		Options: tables,
	}
	survey.AskOne(qs, &tableName)

	db, err := getDB()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var whereClause string
	var useBuilder bool

	qsBuilder := &survey.Confirm{
		Message: "Use condition builder to create filters? (Recommended - No SQL knowledge needed)",
		Default: true,
	}
	survey.AskOne(qsBuilder, &useBuilder)

	if useBuilder {
		whereClause, err = PrintConditionBuilder(db, tableName)
		if err != nil {
			fmt.Printf("Error building conditions: %v\n", err)
			return
		}
	} else {
		columns, _ := getTableColumns(tableName)
		fmt.Printf("\nAvailable columns: %s\n\n", strings.Join(columns, ", "))
		qsWhere := &survey.Input{
			Message: "Enter WHERE condition (or press Enter for all):",
		}
		survey.AskOne(qsWhere, &whereClause)
	}

	var limitStr string
	qsLimit := &survey.Input{
		Message: "Enter rows per page (default 20):",
		Default: "20",
	}
	survey.AskOne(qsLimit, &limitStr)

	limit := 20
	if limitStr != "" {
		fmt.Sscanf(limitStr, "%d", &limit)
	}

	page := 1

	for {
		ctx := context.Background()
		exec := query.NewExecutor(db)

		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM \"%s\"", tableName)
		if whereClause != "" {
			countQuery += " WHERE " + whereClause
		}

		var totalCount int64
		row, _ := db.QueryRow(ctx, countQuery)
		if err := row.(interface{ Scan(...interface{}) error }).Scan(&totalCount); err != nil {
			fmt.Printf("Error counting: %v\n", err)
			return
		}

		if totalCount == 0 {
			color.Yellow.Println("\nNo records match your filter criteria.")
			return
		}

		totalPages := int(totalCount) / limit
		if int(totalCount)%limit > 0 {
			totalPages++
		}

		offset := (page - 1) * limit
		queryStr := fmt.Sprintf("SELECT * FROM \"%s\"", tableName)
		if whereClause != "" {
			queryStr += " WHERE " + whereClause
		}
		queryStr += fmt.Sprintf(" ORDER BY 1 LIMIT %d OFFSET %d", limit, offset)

		results, err := exec.Execute(queryStr)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		remaining := int(totalCount) - (page * limit)
		if remaining < 0 {
			remaining = 0
		}

		color.Cyan.Printf("\n%s\n", strings.Repeat("=", 70))
		color.Bold.Printf("  Filter Results: %s\n", tableName)
		color.Cyan.Printf("%s\n", strings.Repeat("=", 70))
		fmt.Printf("  Page %d of %d (Total: %d records, %d remaining)\n", page, totalPages, totalCount, remaining)
		color.Cyan.Printf("%s\n\n", strings.Repeat("=", 70))
		PrintTable(results)

		remainingPages := totalPages - page
		if remainingPages < 0 {
			remainingPages = 0
		}

		fmt.Println()
		fmt.Printf("Page %d of %d (%d pages remaining)\n", page, totalPages, remainingPages)
		fmt.Println("1. Next Page")
		fmt.Println("2. Previous Page")
		fmt.Println("3. Go to Specific Page")
		fmt.Println("4. Export Current Results")
		fmt.Println("5. Back to Main Menu")

		var choice string
		qsNav := &survey.Select{
			Message: "What would you like to do?",
			Options: []string{"Next Page", "Previous Page", "Go to Specific Page", "Export Current Results", "Back to Main Menu"},
			Default: "Back to Main Menu",
		}
		survey.AskOne(qsNav, &choice)

		switch choice {
		case "Next Page":
			page++
		case "Previous Page":
			if page > 1 {
				page--
			} else {
				fmt.Println("Already on first page!")
			}
		case "Go to Specific Page":
			var pageStr string
			qsPage := &survey.Input{
				Message: "Enter page number:",
			}
			survey.AskOne(qsPage, &pageStr)
			var newPage int
			fmt.Sscanf(pageStr, "%d", &newPage)
			if newPage > 0 && newPage <= totalPages {
				page = newPage
			} else {
				fmt.Printf("Invalid page number. Valid range: 1-%d\n", totalPages)
			}
		case "Export Current Results":
			fmt.Println("\nExporting current results...")
			exportCurrentResults(tableName, results)
		case "Back to Main Menu":
			fmt.Println("\nReturning to main menu...")
			return
		}
	}
}

func exportCurrentResults(tableName string, results []map[string]interface{}) {
	var format string
	qsFormat := &survey.Select{
		Message: "Select export format:",
		Options: []string{"json", "csv"},
		Default: "json",
	}
	survey.AskOne(qsFormat, &format)

	defaultPath := fmt.Sprintf("%s_filtered.%s", tableName, format)
	if downloads := getDownloadsDir(); downloads != "" {
		defaultPath = filepath.Join(downloads, defaultPath)
	}

	var outputPath string
	qsPath := &survey.Input{
		Message: "Enter output file path:",
		Default: defaultPath,
	}
	survey.AskOne(qsPath, &outputPath)

	err := writeExportFile(outputPath, format, results)
	if err != nil {
		fmt.Printf("Error exporting: %v\n", err)
	} else {
		fmt.Printf("\nSuccessfully exported %d records to %s\n", len(results), outputPath)
	}
}

func writeExportFile(path, format string, results []map[string]interface{}) error {
	if len(results) == 0 {
		return os.WriteFile(path, []byte("[]"), 0644)
	}

	if format == "json" {
		data, _ := json.MarshalIndent(results, "", "  ")
		return os.WriteFile(path, data, 0644)
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	cols := make([]string, 0)
	for key := range results[0] {
		cols = append(cols, key)
	}
	writer.Write(cols)

	for _, row := range results {
		record := make([]string, len(cols))
		for i, col := range cols {
			if val, ok := row[col]; ok {
				record[i] = fmt.Sprintf("%v", val)
			}
		}
		writer.Write(record)
	}

	return nil
}

func interactiveTransform() {
	tables, err := getExistingTables()
	if err != nil {
		fmt.Println("Error getting tables:", err)
		return
	}

	if len(tables) == 0 {
		fmt.Println("No tables found in database. Please migrate a file first.")
		return
	}

	var tableName string
	qs := &survey.Select{
		Message: "Select table to transform:",
		Options: tables,
	}
	survey.AskOne(qs, &tableName)

	columns, err := getTableColumns(tableName)
	if err != nil {
		fmt.Println("Error getting columns:", err)
		return
	}

	fmt.Printf("\nAvailable columns: %s\n", strings.Join(columns, ", "))

	var columnsStr string
	qsCol := &survey.Input{
		Message: "Enter columns to select (comma-separated):",
	}
	survey.AskOne(qsCol, &columnsStr)

	selectedCols := strings.Split(columnsStr, ",")
	for i := range selectedCols {
		selectedCols[i] = strings.TrimSpace(selectedCols[i])
	}

	var whereClause string
	qsWhere := &survey.Input{
		Message: "Enter WHERE condition (optional):",
	}
	survey.AskOne(qsWhere, &whereClause)

	fmt.Println("\nRunning transform...")

	db, err := getDB()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	exec := query.NewExecutor(db)

	colList := strings.Join(selectedCols, ", ")
	queryStr := fmt.Sprintf("SELECT %s FROM \"%s\"", colList, tableName)
	if whereClause != "" {
		queryStr += " WHERE " + whereClause
	}

	results, err := exec.Execute(queryStr)
	if err != nil {
		fmt.Printf("Error executing query: %v\n", err)
		return
	}

	if len(results) == 0 {
		fmt.Println("No results found")
		return
	}

	fmt.Printf("\nResults (%d columns, %d rows):\n\n", len(selectedCols), len(results))
	PrintTable(results)

	askExport(results)
}

func interactivePaginate() {
	tables, err := getExistingTables()
	if err != nil {
		fmt.Println("Error getting tables:", err)
		return
	}

	if len(tables) == 0 {
		fmt.Println("No tables found in database. Please migrate a file first.")
		return
	}

	var tableName string
	qs := &survey.Select{
		Message: "Select table to paginate:",
		Options: tables,
	}
	survey.AskOne(qs, &tableName)

	var pageSizeStr string
	qsPageSize := &survey.Input{
		Message: "Enter rows per page (default 20):",
		Default: "20",
	}
	survey.AskOne(qsPageSize, &pageSizeStr)

	pageSize := 20
	if pageSizeStr != "" {
		fmt.Sscanf(pageSizeStr, "%d", &pageSize)
	}

	page := 1

	for {
		ctx := context.Background()
		paginateWithContext(ctx, tableName, page, pageSize)

		fmt.Println("\n=== Navigation ===")
		fmt.Println("1. Next Page")
		fmt.Println("2. Previous Page")
		fmt.Println("3. Go to Specific Page")
		fmt.Println("4. Export Current Results")
		fmt.Println("5. Back to Main Menu")

		var choice string
		qsNav := &survey.Select{
			Message: "What would you like to do?",
			Options: []string{"Next Page", "Previous Page", "Go to Specific Page", "Export Current Results", "Back to Main Menu"},
			Default: "Back to Main Menu",
		}
		survey.AskOne(qsNav, &choice)

		switch choice {
		case "Next Page":
			page++
		case "Previous Page":
			if page > 1 {
				page--
			} else {
				fmt.Println("Already on first page!")
			}
		case "Go to Specific Page":
			var pageStr string
			qsPage := &survey.Input{
				Message: "Enter page number:",
			}
			survey.AskOne(qsPage, &pageStr)
			var newPage int
			fmt.Sscanf(pageStr, "%d", &newPage)
			if newPage > 0 {
				page = newPage
			}
		case "Export Current Results":
			exportTableData(tableName)
		case "Back to Main Menu":
			return
		}
	}
}

func interactiveSchema() {
	tables, err := getExistingTables()
	if err != nil {
		fmt.Println("Error getting tables:", err)
		return
	}

	if len(tables) == 0 {
		fmt.Println("No tables found in database. Please migrate a file first.")
		return
	}

	var tableName string
	qs := &survey.Select{
		Message: "Select table to view schema:",
		Options: tables,
	}
	survey.AskOne(qs, &tableName)

	fmt.Println("\nGetting schema...")
	schemaCmd.Run(nil, []string{tableName})
}

func getDownloadsDir() string {
	home, _ := os.UserHomeDir()
	if home == "" {
		return ""
	}
	downloadsPath := filepath.Join(home, "Downloads")
	if _, err := os.Stat(downloadsPath); err == nil {
		return downloadsPath
	}
	return ""
}

func interactiveExport() {
	tables, err := getExistingTables()
	if err != nil {
		fmt.Println("Error getting tables:", err)
		return
	}

	if len(tables) == 0 {
		fmt.Println("No tables found in database. Please migrate a file first.")
		return
	}

	var tableName string
	qs := &survey.Select{
		Message: "Select table to export:",
		Options: tables,
	}
	survey.AskOne(qs, &tableName)

	var format string
	qsFormat := &survey.Select{
		Message: "Select export format:",
		Options: []string{"json", "csv"},
		Default: "json",
	}
	survey.AskOne(qsFormat, &format)

	defaultPath := "export." + format
	if downloads := getDownloadsDir(); downloads != "" {
		defaultPath = filepath.Join(downloads, defaultPath)
	}

	var outputPath string
	qsPath := &survey.Input{
		Message: "Enter output file path:",
		Default: defaultPath,
	}
	survey.AskOne(qsPath, &outputPath)

	fmt.Println("\nExporting data...")
	exportWithContext(tableName, format, outputPath)
}

func interactiveDelete() {
	tables, err := getExistingTables()
	if err != nil {
		fmt.Println("Error getting tables:", err)
		return
	}

	if len(tables) == 0 {
		fmt.Println("No tables found in database. Please migrate a file first.")
		return
	}

	var tableName string
	qs := &survey.Select{
		Message: "Select table to delete:",
		Options: tables,
	}
	survey.AskOne(qs, &tableName)

	var confirm bool
	qsConfirm := &survey.Confirm{
		Message: fmt.Sprintf("Are you sure you want to delete table '%s'? This cannot be undone.", tableName),
		Default: false,
	}
	survey.AskOne(qsConfirm, &confirm)

	if !confirm {
		fmt.Println("Cancelled.")
		return
	}

	db, err := getDB()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	ctx := context.Background()
	query := fmt.Sprintf("DROP TABLE IF EXISTS \"%s\"", tableName)
	_, err = db.Exec(ctx, query)
	if err != nil {
		fmt.Printf("Error deleting table: %v\n", err)
		return
	}

	color.Green.Printf("\nTable '%s' deleted successfully!\n", tableName)
}

func exportTableData(tableName string) {
	db, err := getDB()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	q := query.New(tableName)
	exec := query.NewExecutor(db)

	queryStr, _ := q.Filter("")

	results, err := exec.Execute(queryStr)
	if err != nil {
		fmt.Printf("Error executing query: %v\n", err)
		return
	}

	if len(results) == 0 {
		fmt.Println("No results to export")
		return
	}

	var format string
	qsFormat := &survey.Select{
		Message: "Select export format:",
		Options: []string{"json", "csv"},
		Default: "json",
	}
	survey.AskOne(qsFormat, &format)

	defaultPath := fmt.Sprintf("%s_export.%s", tableName, format)
	if downloads := getDownloadsDir(); downloads != "" {
		defaultPath = filepath.Join(downloads, defaultPath)
	}

	var outputPath string
	qsPath := &survey.Input{
		Message: "Enter output file path:",
		Default: defaultPath,
	}
	survey.AskOne(qsPath, &outputPath)

	fmt.Println("\nExporting data...")
	exportWithContext(tableName, format, outputPath)
}
