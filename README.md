# Dataset CLI

A powerful, Google-quality CLI tool for processing and querying datasets with PostgreSQL.

![Dataset CLI](https://img.shields.io/badge/version-1.0.0-blue.svg)
![Go](https://img.shields.io/badge/Go-1.21+-00ADD8.svg)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16+-336791.svg)

## Features

- 📊 **Smart Data Import** - Import CSV/JSON files with automatic type detection
- 🔍 **Visual Query Builder** - No SQL knowledge required
- 🎯 **Multiple Output Formats** - Table, JSON, CSV, Markdown, Pretty
- ⚡ **Streaming Import** - Handle large files efficiently
- 🔧 **Data Validation** - Pre-flight checks before import
- 🎨 **Beautiful CLI** - Color-coded output, progress bars, spinners
- 📈 **Statistics** - Table insights and data profiling
- 🐳 **Docker Ready** - One-command setup

## Quick Start

### Option 1: Download Binary

Download the latest release for your platform from [GitHub Releases](https://github.com/your-repo/releases).

### Option 2: Build from Source

```bash
git clone https://github.com/your-repo/dataset-cli.git
cd dataset-cli
go build -o dataset-cli .
```

### Option 3: Docker

```bash
# Start PostgreSQL
make start

# Run CLI
make run

# Or directly
docker compose up -d postgres
docker compose run --rm dataset-cli
```

## Usage

### Interactive Mode

```bash
./dataset-cli
```

### Migrate Data

```bash
# Basic import
./dataset-cli migrate data.csv

# With options
./dataset-cli migrate data.csv \
  --table-name my_data \
  --drop \
  --skip-errors \
  --progress

# Dry run (preview)
./dataset-cli migrate data.csv --dry-run
```

### Filter Data

```bash
# Using condition builder (interactive)
./dataset-cli filter

# Using SQL
./dataset-cli filter users --where "age > 25" --limit 10

# All data
./dataset-cli filter users
```

### Transform Data

```bash
./dataset-cli transform users --columns name,email,age

# With filter
./dataset-cli transform users \
  --columns name,email \
  --where "city = 'New York'"
```

### Export Data

```bash
# JSON (default)
./dataset-cli export users --output data.json

# CSV
./dataset-cli export users --output data.csv --format csv

# Markdown (for docs)
./dataset-cli export users --output data.md --format md

# Pretty (terminal)
./dataset-cli export users --format pretty
```

### View Schema

```bash
./dataset-cli schema users
```

### Statistics

```bash
# Basic stats
./dataset-cli stats users

# With NULL counts
./dataset-cli stats users --show-nulls
```

### Delete Table

```bash
# With confirmation
./dataset-cli delete users

# Force (no confirmation)
./dataset-cli delete users --force
```

### Health Check

```bash
./dataset-cli doctor
```

## Global Flags

```bash
./dataset-cli [command] [flags]

Flags:
  -v, --verbose      Enable verbose output
      --dry-run     Show what would be done
      --no-color    Disable colors
      --config      Config file path
      --host        Database host
      --port        Database port
      --user        Database user
      --password    Database password
      --dbname      Database name
```

## Configuration

Create `~/.dataset-cli/.env` or `~/.dataset-cli/config.yaml`:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=dataset
DB_SSLMODE=disable
```

## Shell Completion

### Bash

```bash
# Add to ~/.bashrc
source /path/to/completions/bash

# Or use dataset-cli
./dataset-cli completion bash > /etc/bash_completion.d/dataset-cli
```

### Zsh

```bash
# Add to ~/.zshrc
source /path/to/completions/zsh

# Or use dataset-cli
./dataset-cli completion zsh > "${fpath[1]}/_dataset-cli"
```

## Architecture

```
dataset-cli/
├── cmd/                    # CLI commands
│   ├── migrate.go         # Import data
│   ├── filter.go           # Filter queries
│   ├── transform.go        # Column selection
│   ├── export.go           # Export data
│   ├── schema.go           # View schema
│   ├── stats.go            # Statistics
│   ├── delete.go           # Delete tables
│   ├── doctor.go           # Health checks
│   ├── wizard.go           # Interactive mode
│   ├── condition_builder.go # Query builder UI
│   ├── table.go            # Table formatting
│   └── format.go           # Output formatters
├── internal/
│   ├── analyzer/           # Type detection
│   ├── database/           # DB connection
│   ├── query/              # Query builder
│   ├── reader/             # File reading
│   ├── validator/          # Pre-flight checks
│   ├── progress/           # Progress bars
│   └── errors/             # Error handling
├── completions/            # Shell completions
├── Dockerfile
├── docker-compose.yml
└── Makefile
```

## Development

```bash
# Build
make build

# Run
make run

# Test
go test ./...

# Clean
make clean
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests
5. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) for details.

## Support

- 📖 [Documentation](docs/)
- 🐛 [Issues](https://github.com/your-repo/issues)
- 💬 [Discussions](https://github.com/your-repo/discussions)

---

Made with ❤️ using Go
