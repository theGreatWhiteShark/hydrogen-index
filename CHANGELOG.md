# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/).

## [Unreleased]

### Added

- `--exclude` / `-e` flag to blacklist folders from scanning (useful for CI
  and submodule setups to exclude test artifacts).
- `folderName` field for drumkit entries in index.json (extracted from
  `.h2drumkit` tar archive structure).
- `--base-url`, `--provider`, `--repo`, and `--branch` flags for constructing
  permalink URLs in `index.json`.
- Provider presets for GitHub and GitLab raw-file URL generation.

### Changed

- Calculation of top-level HASH was changed in order to the C++ counter part to
  properly parse it.

### Fixed

- Field "components" does now enumerate the instrument components instead of the
  legacy drumkit components (aux channels to route instrument components).

### Removed

- Support for standalone `drumkit.xml` files. Only `.h2drumkit` archives are now
  recognized.

## [0.1.0] - 2026-03-27

### Added

- `scan` command to discover and index Hydrogen artifacts in a directory tree
- `validate` command to check index.json files against the JSON schema
- Support for `.h2drumkit` (tar), `.h2pattern`, `.h2song`, and standalone `drumkit.xml` files
- Legacy format support spanning Hydrogen versions 0.9.3 through 2.0.0
- Automatic git repository root detection when no directory is specified
- SHA-256 hash computation for artifacts and self-hashing of the index file
- JSON Schema validation for generated index files
