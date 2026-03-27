# hydrogen-index

A CLI tool that scans a directory tree for Hydrogen music application artifacts,
parses their XML metadata, and produces a structured `index.json` summary file.
It also includes a `validate` command to check whether a JSON file conforms to
the index schema.

## Installation

```
go install github.com/hydrogen-music/hydrogen-index@latest
```

Or build from source:

```
git clone https://github.com/hydrogen-music/hydrogen-index.git
cd hydrogen-index
go build -o hydrogen-index .
```

## Usage

### Scan (default command)

```
# Scan from git repo root (auto-detected)
hydrogen-index

# Scan a specific directory
hydrogen-index scan -d /path/to/artifacts

# Specify output file
hydrogen-index scan -d /path/to/artifacts -o custom-index.json
```

### Validate

```
hydrogen-index validate index.json
```

### Version

```
hydrogen-index --version
```

## Supported Artifact Formats

- `.h2drumkit` — Hydrogen drumkit archives (tar format containing drumkit.xml)
- `.h2pattern` — Hydrogen pattern files
- `.h2song` — Hydrogen song files
- `drumkit.xml` — Standalone drumkit XML files

Supports format versions from 0.9.3 through 2.0.0.

## Output Format

The `scan` command writes an `index.json` file containing metadata extracted
from all discovered artifacts, along with SHA-256 hashes for each artifact and
a self-hash of the index file itself. See
`docs/decisions/0002-index-file-format.md` for the full specification.

## Development

```
go test ./...
```

## License

GPLv3 — see LICENSE file.
