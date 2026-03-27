# hydrogen-index

`hydrogen-index` is a CLI tool designed to traverse a directory structure (by default the root of a git repository), discover Hydrogen artifacts (drumkits, songs, patterns), and generate an `index.json` summarizing them. This allows users to easily host and share their custom content for online import into the Hydrogen application.

## Features

- **scan (default):** Scans the file system for Hydrogen artifacts and creates an `index.json` file.
  - Automatically finds the root of the current git repository to start the scan, or takes a specific directory via the `-d` flag.
  - Handles legacy formats and automatically extracts `.h2drumkit` tar archives to parse the `drumkit.xml`.
  - Generates a JSON file (`index.json` by default, customizable via `-o`) summarizing the metadata.
- **validate:** Checks whether a provided input file matches the expected `index.json` format, using a JSON schema.

## Usage

```bash
# General scan in a git repository
hydrogen-index scan

# Scan a specific directory
hydrogen-index scan -d /path/to/artifacts

# Specify a custom output file
hydrogen-index scan -o /path/to/custom-index.json

# Validate an index file
hydrogen-index validate input-index.json
```

## Development

This project is written in Go and uses `github.com/spf13/cobra` for its CLI interface. It is developed using Test-Driven Development (TDD). 

### Prerequisites

- Go 1.21+

### Building

```bash
go build -o hydrogen-index
```

### Testing

```bash
go test ./...
```
