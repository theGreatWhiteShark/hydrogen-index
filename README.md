# hydrogen-index

`hydrogen-index` is a small Go CLI that scans Hydrogen artifacts in a folder tree and writes a JSON index suitable for online import workflows.

## Features

- uses `cobra` for CLI argument parsing
- supports `--help` and `--version`
- defaults to `scan` when invoked without a subcommand
- scans the current git repository root automatically
- can scan an explicit directory with `-d`
- writes `index.json` to the current working directory by default
- supports a custom output path with `-o`
- derives artifact URLs from a required base URL prefix when provided via `--base-url`
- validates generated or hand-written indexes against a bundled JSON Schema via `validate`

## Supported artifacts

- `.h2song`
- `.h2pattern`
- `.h2drumkit` tar archives containing `drumkit.xml`
- legacy `drumkit.xml` layouts
- current and legacy XML layouts found in `res/hydrogen-artifacts`

## Usage

```bash
hydrogen-index --base-url https://example.com/repository
hydrogen-index scan --base-url https://example.com/repository
hydrogen-index scan -d ./content -o ./public/index.json --base-url https://example.com/repository
hydrogen-index validate ./public/index.json
```

## Commands

### `scan`

When no `-d` flag is provided, `hydrogen-index` walks upward from the current working directory until it finds the enclosing git repository root and scans that tree recursively.

Options:

- `-d, --directory`: scan this directory directly instead of discovering a git repository root
- `-o, --output`: write the JSON index to this path instead of `./index.json`
- `--base-url`: prepend this base URL to repository-relative artifact paths when populating each artifact `url`

Both `-d` and `-o` accept absolute and relative paths.

### `validate`

```bash
hydrogen-index validate path/to/index.json
```

This checks that the input JSON matches the bundled schema derived from MADR 0002 and the reference document in `res/references-index.json`.

## Output

The generated file follows the structure outlined in `docs/decisions/0002-index-file-format.md` and resembles `res/references-index.json`.

Top-level fields:

- `version`
- `created`
- `patternCount`
- `songCount`
- `drumkitCount`
- `patterns`
- `songs`
- `drumkits`
- `hash`

Artifact URLs are emitted as repository-relative paths when `--base-url` is omitted, or as base-URL-prefixed permalinks when `--base-url` is supplied.
