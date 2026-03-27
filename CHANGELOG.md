# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/).

## [0.1.0] - 2026-03-27

### Added

- `scan` command to discover and index Hydrogen artifacts in a directory tree
- `validate` command to check index.json files against the JSON schema
- Support for `.h2drumkit` (tar), `.h2pattern`, `.h2song`, and standalone `drumkit.xml` files
- Legacy format support spanning Hydrogen versions 0.9.3 through 2.0.0
- Automatic git repository root detection when no directory is specified
- SHA-256 hash computation for artifacts and self-hashing of the index file
- JSON Schema validation for generated index files
