package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <table>",
	Short: "Delete a table from the database",
	Long: `Delete (drop) a table from the database. This will permanently remove the table and all its data.

Example:
  dataset-cli delete users
  dataset-cli delete my_table --force`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		tableName := args[0]

		force, _ := cmd.Flags().GetBool("force")
		if !force {
			fmt.Printf("Are you sure you want to delete table '%s'? This action cannot be undone. (y/N): ", tableName)
			var confirm string
			fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Println("Cancelled.")
				return
			}
		}

		db, err := getDB()
		if err != nil {
			PrintError("%v", err)
			os.Exit(1)
		}
		ctx := context.Background()
		exists, err := db.TableExists(ctx, tableName)
		if err != nil {
			PrintError("Error checking table: %v", err)
			os.Exit(1)
		}
		if !exists {
			PrintError("Table '%s' does not exist", tableName)
			os.Exit(1)
		}

		query := fmt.Sprintf("DROP TABLE IF EXISTS \"%s\"", tableName)
		_, err = db.Exec(ctx, query)
		if err != nil {
			PrintError("Failed to delete table: %v", err)
			os.Exit(1)
		}

		PrintSuccess("Table '%s' deleted successfully", tableName)
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().Bool("force", false, "Skip confirmation prompt")
}
