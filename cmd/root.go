package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hydrogen-music/hydrogen-index/internal/index"
	"github.com/hydrogen-music/hydrogen-index/internal/model"
	"github.com/hydrogen-music/hydrogen-index/internal/scanner"
	"github.com/spf13/cobra"
)

// Execute builds the command tree and runs it. Returns any error so main can
// exit with a non-zero status without printing a duplicate error line.
func Execute() error {
	rootCmd := buildRootCmd()
	rootCmd.AddCommand(buildScanCmd())
	rootCmd.AddCommand(buildValidateCmd())
	return rootCmd.Execute()
}

func buildRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hydrogen-index",
		Short: "Index and validate Hydrogen music artifacts",
		// RunE delegates to scan so `hydrogen-index` with no subcommand behaves
		// identically to `hydrogen-index scan`.
		RunE:    runScan,
		Version: model.Version,
	}

	// Custom template keeps the version output consistent with the spec:
	// "hydrogen-index version 0.1.0"
	cmd.SetVersionTemplate("hydrogen-index version {{.Version}}\n")

	addScanFlags(cmd)
	return cmd
}

// scanFlags holds the resolved flag values for the scan command. Defined at
// package level so both the root RunE and the explicit scan subcommand share
// the same addScanFlags helper without coupling them.
type scanFlags struct {
	directory string
	output    string
}

func addScanFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("directory", "d", "", "directory to scan (default: git repo root)")
	cmd.Flags().StringP("output", "o", "", "output file path (default: index.json in scan directory)")
}

func buildScanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan a directory and produce index.json",
		RunE:  runScan,
	}
	addScanFlags(cmd)
	return cmd
}

// runScan is the shared RunE for both the root command and the scan subcommand.
func runScan(cmd *cobra.Command, args []string) error {
	flags, err := resolveScanFlags(cmd)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Scanning %s...\n", flags.directory)

	artifacts, scanErrs := scanner.Scan(flags.directory)
	for _, e := range scanErrs {
		fmt.Fprintf(os.Stderr, "warning: %v\n", e)
	}

	idx, err := index.Build(artifacts)
	if err != nil {
		return fmt.Errorf("building index: %w", err)
	}

	data, err := index.Finalize(idx)
	if err != nil {
		return fmt.Errorf("finalizing index: %w", err)
	}

	if err := os.WriteFile(flags.output, data, 0644); err != nil {
		return fmt.Errorf("writing %s: %w", flags.output, err)
	}

	fmt.Fprintf(os.Stderr, "Wrote %s (%d patterns, %d songs, %d drumkits)\n",
		flags.output, idx.PatternCount, idx.SongCount, idx.DrumkitCount)
	return nil
}

// resolveScanFlags derives final directory and output paths from flags,
// falling back to git-root detection and sensible defaults.
func resolveScanFlags(cmd *cobra.Command) (scanFlags, error) {
	dir, _ := cmd.Flags().GetString("directory")
	out, _ := cmd.Flags().GetString("output")

	if dir == "" {
		root, err := findGitRoot()
		if err != nil {
			return scanFlags{}, fmt.Errorf("--directory not specified and no git repository found: %w", err)
		}
		dir = root
	}

	if out == "" {
		out = filepath.Join(dir, "index.json")
	}

	return scanFlags{directory: dir, output: out}, nil
}

// findGitRoot walks upward from the current working directory until it finds a
// directory containing a .git entry, or returns an error if it reaches the
// filesystem root without finding one.
func findGitRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("could not determine current directory: %w", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root with no .git found.
			return "", errors.New("no git repository found in any parent directory")
		}
		dir = parent
	}
}
