# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added
- JSON Schema validation for index files in `pkg/validator`
- Indexer orchestrator for building Hydrogen artifact indices
- Integration between `scan` command and indexer
- Support for automatic Git root detection when no directory is specified
- SHA-256 hashing of index files for integrity verification
- XML parser for Hydrogen artifacts (.h2pattern, .h2song, .h2drumkit)
- File scanner for locating Hydrogen artifacts and extracting drumkit metadata
- JSON domain structs and unmarshaling logic for the index file format
- CLI structure using cobra
- `scan` command with `--dir` and `--out` flags
- `validate` command with positional file argument
- `version` command
- Default behavior to run `scan` when no subcommand is provided
