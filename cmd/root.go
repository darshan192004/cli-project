package cmd

import (
	"context"
	"fmt"
	"os"

	"dataset-cli/internal/config"
	"dataset-cli/internal/database"
	"github.com/spf13/cobra"
)

var cfgFile string
var db *database.DB
var selectColumns []string

var rootCmd = &cobra.Command{
	Use:   "dataset-cli",
	Short: "CLI tool for processing and querying datasets",
	Long: `Dataset CLI - A powerful tool for processing datasets

Run without arguments to start interactive mode, or use:
  migrate    - Import CSV/JSON files into PostgreSQL
  filter     - Filter data with WHERE conditions
  transform  - Select specific columns
  paginate   - Paginated queries
  schema     - Show table schema
  export     - Export query results to file

For more information, use --help with any command.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.dataset-cli.yaml)")
	rootCmd.PersistentFlags().String("host", "", "Database host (overrides config)")
	rootCmd.PersistentFlags().Int("port", 0, "Database port (overrides config)")
	rootCmd.PersistentFlags().String("user", "", "Database user (overrides config)")
	rootCmd.PersistentFlags().String("password", "", "Database password (overrides config)")
	rootCmd.PersistentFlags().String("dbname", "", "Database name (overrides config)")
}

func loadConfig(host string, port int, user, password, dbname string) (*config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	if host != "" {
		cfg.Database.Host = host
	}
	if port != 0 {
		cfg.Database.Port = port
	}
	if user != "" {
		cfg.Database.User = user
	}
	if password != "" {
		cfg.Database.Password = password
	}
	if dbname != "" {
		cfg.Database.DBName = dbname
	}

	return cfg, nil
}

func getDB() (*database.DB, error) {
	if db != nil {
		err := db.Pool.Ping(context.Background())
		if err == nil {
			return db, nil
		}
	}

	host, _ := rootCmd.Flags().GetString("host")
	port, _ := rootCmd.Flags().GetInt("port")
	user, _ := rootCmd.Flags().GetString("user")
	password, _ := rootCmd.Flags().GetString("password")
	dbname, _ := rootCmd.Flags().GetString("dbname")

	cfg, err := loadConfig(host, port, user, password, dbname)
	if err != nil {
		return nil, err
	}

	newDB, err := database.Connect(&cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db = newDB
	return db, nil
}
