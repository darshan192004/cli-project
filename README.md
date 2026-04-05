# Dataset CLI

A powerful CLI tool for processing and querying datasets with support for SQLite, PostgreSQL, and TursoDB (cloud SQLite).

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go](https://img.shields.io/badge/Go-1.26+-00ADD8.svg)
![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-green.svg)

## Features

- **Multiple Databases** - Works with SQLite (default), PostgreSQL, and TursoDB cloud
- **Smart Data Import** - Import CSV/JSON files with automatic type detection
- **Beautiful CLI** - Color-coded output, progress bars, interactive prompts
- **No SQL Required** - Interactive query builder for non-technical users
- **Streaming Import** - Handle large files efficiently with low memory usage
- **Data Validation** - Pre-flight checks before import
- **Multiple Output Formats** - Table, JSON, CSV, Markdown, Pretty print

## Installation

### npm (Recommended)

```bash
npm install -g dataset-cli
```

Works on Windows, macOS, and Linux. Binary is downloaded automatically on first run.

### Download Binary

Download the latest release for your platform from [GitHub Releases](https://github.com/darshan192004/cli-project/releases).

### Docker

```bash
docker pull darshan192004/dataset-cli
docker run --rm darshan192004/dataset-cli --help
```

### Build from Source

```bash
git clone https://github.com/darshan192004/cli-project.git
cd cli-project
go build -o dataset-cli .
```

## Quick Start

```bash
# Start interactive mode
dataset-cli

# Import a CSV file
dataset-cli migrate data.csv

# Filter data with SQL-like syntax
dataset-cli filter users --where "age > 25"

# Export to JSON
dataset-cli export users --output data.json

# Check system health
dataset-cli doctor
```

## Commands

### `migrate` - Import Data

```bash
# Basic import (creates table from filename)
dataset-cli migrate data.csv

# Specify table name
dataset-cli migrate data.csv --table-name users

# With options
dataset-cli migrate data.csv \
  --table-name users \
  --drop \           # Drop existing table first
  --skip-errors \    # Continue on errors
  --progress         # Show progress bar

# Use PostgreSQL
dataset-cli migrate data.csv --postgres

# Use TursoDB Cloud
dataset-cli migrate data.csv --cloud
```

### `filter` - Query Data

```bash
# Interactive mode
dataset-cli filter

# SQL-like query
dataset-cli filter users --where "age > 25 AND city = 'NYC'" --limit 10

# Simple query
dataset-cli filter users
```

### `transform` - Select Columns

```bash
# Select specific columns
dataset-cli transform users --columns name,email,age

# With filter
dataset-cli transform users \
  --columns name,email \
  --where "city = 'New York'"
```

### `export` - Export Data

```bash
# JSON (default)
dataset-cli export users --output data.json

# CSV
dataset-cli export users --output data.csv --format csv

# Markdown (great for docs)
dataset-cli export users --output data.md --format md

# Pretty print to terminal
dataset-cli export users --format pretty
```

### `schema` - View Table Schema

```bash
dataset-cli schema users
```

### `stats` - Table Statistics

```bash
# Basic stats
dataset-cli stats users

# Include NULL counts
dataset-cli stats users --show-nulls
```

### `aggregate` - Aggregate Functions

```bash
# Count records
dataset-cli aggregate users --operation count

# Sum a column
dataset-cli aggregate orders --operation sum --column amount

# Average, min, max
dataset-cli aggregate users --operation avg --column age
```

### `backup` & `restore` - Data Backup

```bash
# Backup a table
dataset-cli backup users --output backup.json

# Restore from backup
dataset-cli restore users --input backup.json
```

### `doctor` - System Diagnostics

```bash
dataset-cli doctor
```

## Global Flags

```bash
dataset-cli [command] [flags]

Flags:
  -v, --verbose           Enable verbose output
      --dry-run           Show what would be done without executing
      --no-color          Disable color output
      --config string     Config file path (default ~/.dataset-cli.yaml)
      --postgres          Use PostgreSQL instead of SQLite
      --cloud             Use TursoDB cloud (requires LIBSQL_URL env var)
      --host string       Database host (overrides config)
      --port int          Database port (overrides config)
      --user string       Database user (overrides config)
      --password string   Database password (overrides config)
      --dbname string     Database name (overrides config)
```

## Configuration

Create `~/.dataset-cli.yaml`:

```yaml
database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  dbname: dataset
  sslmode: disable
```

Or use environment variables:

```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=secret
export DB_NAME=dataset
export LIBSQL_URL=libsql://your-db.turso.io?authToken=your-token
```

## Database Support

| Database  | CGO Required | Notes                           |
|-----------|--------------|--------------------------------|
| SQLite    | No           | Default, stores at ~/.dataset-cli/ |
| PostgreSQL| No           | Use `--postgres` flag           |
| TursoDB   | No           | Cloud SQLite, use `--cloud` flag |

## Shell Completion

```bash
# Bash
dataset-cli completion bash >> ~/.bashrc

# Zsh
dataset-cli completion zsh >> ~/.zshrc

# Fish
dataset-cli completion fish > ~/.config/fish/completions/dataset-cli.fish
```

## Development

```bash
# Build
make build

# Run tests
go test ./...

# Run with PostgreSQL
make start    # Start PostgreSQL
make run      # Run CLI
```

## License

MIT License - see [LICENSE](LICENSE) for details.

## Links

- [GitHub Repository](https://github.com/darshan192004/cli-project)
- [Issue Tracker](https://github.com/darshan192004/cli-project/issues)
- [npm Package](https://www.npmjs.com/package/dataset-cli)

---

Made with ❤️ using Go
