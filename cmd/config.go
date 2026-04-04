package cmd

import (
	"fmt"
	"os"

	"dataset-cli/internal/config"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  `View, set, init, and validate configuration for dataset-cli`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			PrintError("Failed to load config: %v", err)
			os.Exit(1)
		}

		color.Bold.Println("\nCurrent Configuration:")
		color.Bold.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		fmt.Printf("\n  %s: %s\n", color.Bold.Sprint("Database Host"), cfg.Database.Host)
		fmt.Printf("  %s: %d\n", color.Bold.Sprint("Database Port"), cfg.Database.Port)
		fmt.Printf("  %s: %s\n", color.Bold.Sprint("Database Name"), cfg.Database.DBName)
		fmt.Printf("  %s: %s\n", color.Bold.Sprint("Database User"), cfg.Database.User)
		fmt.Printf("  %s: %s\n", color.Bold.Sprint("SSL Mode"), cfg.Database.SSLMode)

		configPath := viper.ConfigFileUsed()
		if configPath != "" {
			fmt.Printf("\n  %s: %s\n", color.Bold.Sprint("Config File"), configPath)
		} else {
			fmt.Printf("\n  %s: %s\n", color.Bold.Sprint("Config File"), color.Yellow.Sprint("Not found (using defaults)"))
		}

		color.Bold.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]

		validKeys := map[string]string{
			"host":     "database.host",
			"port":     "database.port",
			"user":     "database.user",
			"password": "database.password",
			"dbname":   "database.dbname",
			"sslmode":  "database.sslmode",
		}

		viperKey, ok := validKeys[key]
		if !ok {
			color.Red.Printf("Invalid key '%s'. Valid keys: host, port, user, password, dbname, sslmode\n", key)
			os.Exit(1)
		}

		home, _ := os.UserHomeDir()
		configDir := home + "/.dataset-cli"
		configPath := configDir + "/config.yaml"

		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			_ = os.MkdirAll(configDir, 0755)
		}

		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath(configDir)

		if _, err := os.Stat(configPath); err == nil {
			_ = viper.ReadInConfig()
		}

		viper.Set(viperKey, value)

		if err := viper.WriteConfigAs(configPath); err != nil {
			PrintError("Failed to write config: %v", err)
			os.Exit(1)
		}

		color.Green.Printf("Set %s = %s\n", key, value)
		color.Cyan.Printf("Saved to %s\n", configPath)
	},
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a default configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		home, err := os.UserHomeDir()
		if err != nil {
			PrintError("Failed to get home directory: %v", err)
			os.Exit(1)
		}

		configDir := home + "/.dataset-cli"
		configPath := configDir + "/config.yaml"

		if _, err := os.Stat(configPath); err == nil {
			color.Yellow.Printf("Config already exists at %s\n", configPath)
			color.Cyan.Println("Use 'dataset-cli config set <key> <value>' to update values")
			return
		}

		if err := os.MkdirAll(configDir, 0755); err != nil {
			PrintError("Failed to create config directory: %v", err)
			os.Exit(1)
		}

		cfg := `database:
  host: localhost
  port: 5432
  user: postgres
  password: your_password_here
  dbname: dataset
  sslmode: disable
`

		if err := os.WriteFile(configPath, []byte(cfg), 0644); err != nil {
			PrintError("Failed to create config file: %v", err)
			os.Exit(1)
		}

		color.Green.Printf("Created config file at %s\n", configPath)
		color.Cyan.Println("Please edit the file and update the password field")
	},
}

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			color.Red.Println("✗ Failed to load config")
			PrintError("%v", err)
			os.Exit(1)
		}

		color.Bold.Println("\nValidating Configuration...")
		color.Bold.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		hasErrors := false

		if cfg.Database.Host == "" {
			color.Red.Println("✗ database.host is required")
			hasErrors = true
		} else {
			color.Green.Printf("✓ database.host: %s\n", cfg.Database.Host)
		}

		if cfg.Database.Port <= 0 || cfg.Database.Port > 65535 {
			color.Red.Println("✗ database.port must be between 1 and 65535")
			hasErrors = true
		} else {
			color.Green.Printf("✓ database.port: %d\n", cfg.Database.Port)
		}

		if cfg.Database.User == "" {
			color.Red.Println("✗ database.user is required")
			hasErrors = true
		} else {
			color.Green.Printf("✓ database.user: %s\n", cfg.Database.User)
		}

		if cfg.Database.DBName == "" {
			color.Red.Println("✗ database.dbname is required")
			hasErrors = true
		} else {
			color.Green.Printf("✓ database.dbname: %s\n", cfg.Database.DBName)
		}

		validSSL := map[string]bool{"disable": true, "require": true, "verify-ca": true, "verify-full": true}
		if !validSSL[cfg.Database.SSLMode] {
			color.Red.Printf("✗ database.sslmode '%s' is invalid (use: disable, require, verify-ca, verify-full)\n", cfg.Database.SSLMode)
			hasErrors = true
		} else {
			color.Green.Printf("✓ database.sslmode: %s\n", cfg.Database.SSLMode)
		}

		color.Bold.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		if hasErrors {
			color.Red.Println("\n✗ Configuration validation failed")
			os.Exit(1)
		}

		color.Green.Println("\n✓ Configuration is valid")
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configValidateCmd)
}
