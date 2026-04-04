package cmd

import (
	"fmt"
	"os"

	"dataset-cli/internal/config"
	"dataset-cli/internal/database"
	"github.com/spf13/cobra"
)

var (
	cfgFile       string
	db            database.Backend
	selectColumns []string
	Verbose       bool
	DryRun        bool
	NoColor       bool
	UsePostgres   bool
	UseCloud      bool
)

var rootCmd = &cobra.Command{
	Use:   "dataset-cli",
	Short: "CLI tool for processing and querying datasets",
	Long: `Dataset CLI - A powerful tool for processing datasets

Run without arguments to start interactive mode, or use:
  migrate    - Import CSV/JSON files into database
  filter     - Filter data with WHERE conditions
  transform  - Select specific columns
  paginate   - Paginated queries
  schema     - Show table schema
  export     - Export query results to file
  doctor     - Check system health and diagnose issues

Storage Backends:
  - SQLite (default): Local storage at ~/.dataset-cli/dataset.db
  - PostgreSQL: Use --postgres flag
  - Cloud: Use --cloud flag (TursoDB)

Examples:
  dataset-cli migrate data.csv --progress
  dataset-cli filter users --where "age > 25"
  dataset-cli doctor
  dataset-cli migrate data.csv --postgres  # Use PostgreSQL
  dataset-cli migrate data.csv --cloud     # Use TursoDB cloud

For more information, use --help with any command.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if NoColor {
			_ = os.Setenv("NO_COLOR", "1")
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.dataset-cli.yaml)")
	rootCmd.PersistentFlags().String("host", "", "Database host (overrides config for PostgreSQL)")
	rootCmd.PersistentFlags().Int("port", 0, "Database port (overrides config for PostgreSQL)")
	rootCmd.PersistentFlags().String("user", "", "Database user (overrides config for PostgreSQL)")
	rootCmd.PersistentFlags().String("password", "", "Database password (overrides config for PostgreSQL)")
	rootCmd.PersistentFlags().String("dbname", "", "Database name (overrides config for PostgreSQL)")
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&DryRun, "dry-run", false, "Show what would be done without executing")
	rootCmd.PersistentFlags().BoolVar(&NoColor, "no-color", false, "Disable color output")
	rootCmd.PersistentFlags().BoolVar(&UsePostgres, "postgres", false, "Use PostgreSQL instead of SQLite")
	rootCmd.PersistentFlags().BoolVar(&UseCloud, "cloud", false, "Use TursoDB cloud (requires LIBSQL_URL env var)")
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

func getDB() (database.Backend, error) {
	if db != nil {
		return db, nil
	}

	var err error
	if UseCloud {
		db, err = database.NewBackend(database.BackendLibSQL, nil)
	} else if UsePostgres {
		host, _ := rootCmd.Flags().GetString("host")
		port, _ := rootCmd.Flags().GetInt("port")
		user, _ := rootCmd.Flags().GetString("user")
		password, _ := rootCmd.Flags().GetString("password")
		dbname, _ := rootCmd.Flags().GetString("dbname")

		cfg, loadErr := loadConfig(host, port, user, password, dbname)
		if loadErr != nil {
			return nil, loadErr
		}

		db, err = database.NewBackend(database.BackendPostgres, cfg)
	} else {
		db, err = database.NewBackend(database.BackendSQLite, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}

func Debug(format string, args ...interface{}) {
	if Verbose {
		fmt.Printf("[DEBUG] "+format+"\n", args...)
	}
}

func Log(format string, args ...interface{}) {
	if Verbose {
		fmt.Printf(format+"\n", args...)
	}
}
