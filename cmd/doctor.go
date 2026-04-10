package cmd

import (
	"context"
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/gookit/color"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check system health and diagnose issues",
	Long: `Run diagnostics to verify your setup is correct.
Checks:
- Database connectivity
- Configuration
- System resources
- File permissions`,
	Run: func(cmd *cobra.Command, args []string) {
		runDiagnostics()
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

func runDiagnostics() {
	color.Bold.Printf("\n")
	color.Bold.Printf("  %s\n", strings.Repeat("=", 50))
	color.Cyan.Printf("       Dataset CLI - Diagnostics")
	color.Bold.Printf("  %s\n", strings.Repeat("=", 50))
	color.Bold.Printf("\n")

	allPassed := true

	color.Bold.Printf("\n[%s] System Information\n", color.Green.Sprint("1/4"))
	color.Gray.Printf("  %s\n", strings.Repeat("-", 40))

	fmt.Printf("  OS: ")
	color.Green.Printf("%s/%s\n", runtime.GOOS, runtime.GOARCH)

	fmt.Printf("  Go Version: ")
	color.Green.Printf("%s\n", runtime.Version())

	fmt.Printf("  CPU Cores: ")
	color.Green.Printf("%d\n", runtime.NumCPU())

	fmt.Printf("  Terminal: ")
	term := os.Getenv("TERM")
	if term == "" {
		term = "unknown"
	}
	color.Green.Printf("%s\n", term)

	fmt.Printf("  Color Support: ")
	if os.Getenv("FORCE_COLOR") != "" || term != "dumb" {
		color.Green.Printf("Enabled\n")
	} else {
		color.Yellow.Printf("Limited (set FORCE_COLOR=1 for full colors)\n")
	}

	color.Bold.Printf("\n[%s] Configuration\n", color.Green.Sprint("2/4"))
	color.Gray.Printf("  %s\n", strings.Repeat("-", 40))

	cfg, err := loadConfig("", 0, "", "", "")
	if err != nil {
		color.Red.Printf("  ✗ Failed to load configuration: %v\n", err)
		allPassed = false
	} else {
		fmt.Printf("  Database Host: ")
		color.Green.Printf("%s\n", cfg.Database.Host)

		fmt.Printf("  Database Port: ")
		color.Green.Printf("%d\n", cfg.Database.Port)

		fmt.Printf("  Database Name: ")
		color.Green.Printf("%s\n", cfg.Database.DBName)

		fmt.Printf("  SSL Mode: ")
		color.Green.Printf("%s\n", cfg.Database.SSLMode)
	}

	color.Bold.Printf("\n[%s] Database Connection\n", color.Green.Sprint("3/4"))
	color.Gray.Printf("  %s\n", strings.Repeat("-", 40))

	db, err := getDB()
	if err != nil {
		color.Red.Printf("  ✗ Cannot connect to database\n")
		color.Gray.Printf("     Error: %v\n", err)
		color.Yellow.Printf("     💡 Check if PostgreSQL is running and credentials are correct\n")
		allPassed = false
	} else {

		fmt.Printf("  Connection: ")
		color.Green.Printf("OK\n")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var version string
		row, err := db.QueryRow(ctx, "SELECT version()")
		if err == nil && row != nil {
			if err := row.(interface{ Scan(...interface{}) error }).Scan(&version); err != nil {
				color.Yellow.Printf("  Version: Unable to determine\n")
			} else {
				shortVersion := strings.Split(version, " ")[0:2]
				color.Green.Printf("  Version: %s\n", strings.Join(shortVersion, " "))
			}
		} else {
			color.Yellow.Printf("  Version: Unable to determine\n")
		}

		var dbSize int64
		row, err = db.QueryRow(ctx, "SELECT pg_database_size(current_database())")
		if err == nil && row != nil {
			if err := row.(interface{ Scan(...interface{}) error }).Scan(&dbSize); err == nil {
				mb := float64(dbSize) / (1024 * 1024)
				color.Green.Printf("  Database Size: %.1f MB\n", mb)
			}
		}

		var tableCount int
		row, err = db.QueryRow(ctx, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'")
		if err == nil && row != nil {
			if err := row.(interface{ Scan(...interface{}) error }).Scan(&tableCount); err == nil {
				color.Green.Printf("  Tables: %d\n", tableCount)
			}
		}
	}

	color.Bold.Printf("\n[%s] Permissions & Environment\n", color.Green.Sprint("4/4"))
	color.Gray.Printf("  %s\n", strings.Repeat("-", 40))

	fmt.Printf("  Config Directory: ")
	home, _ := os.UserHomeDir()
	configDir := home + "/.dataset-cli"
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		color.Yellow.Printf("Does not exist (will be created)\n")
	} else {
		color.Green.Printf("Exists\n")
	}

	fmt.Printf("  Current Directory: ")
	cwd, _ := os.Getwd()
	color.Green.Printf("%s\n", cwd)

	fmt.Printf("  Network: ")
	if netPtr, err := net.LookupHost("localhost"); err != nil || len(netPtr) == 0 {
		color.Yellow.Printf("localhost not resolvable\n")
	} else {
		color.Green.Printf("OK\n")
	}

	fmt.Printf("  DNS: ")
	if ips, err := net.LookupIP("google.com"); err != nil || len(ips) == 0 {
		color.Yellow.Printf("Cannot reach external hosts\n")
	} else {
		color.Green.Printf("OK\n")
	}

	color.Bold.Printf("\n%s\n", strings.Repeat("=", 52))

	if allPassed {
		color.Bold.Printf("  ")
		color.Green.Printf("✓ All checks passed!\n")
		color.Green.Printf("  Your setup is ready to use.\n")
	} else {
		color.Bold.Printf("  ")
		color.Red.Printf("✗ Some checks failed.\n")
		color.Yellow.Printf("  Please review the warnings above.\n")
	}

	color.Bold.Printf("%s\n", strings.Repeat("=", 52))
}
