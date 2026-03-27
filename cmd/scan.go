package cmd

import (
	"fmt"
	"io"

	"github.com/hydrogen-music/hydrogen-index/pkg/indexer"
	"github.com/hydrogen-music/hydrogen-index/pkg/scanner"
	"github.com/spf13/cobra"
)

type ScanOptions struct {
	Dir string
	Out string
}

func NewScanCmd() *cobra.Command {
	opts := &ScanOptions{}

	scanCmd := &cobra.Command{
		Use:   "scan",
		Short: "Traverse file system and create an index.json from Hydrogen artifacts",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runScan(cmd.OutOrStdout(), opts)
		},
	}

	scanCmd.Flags().StringVarP(&opts.Dir, "dir", "d", "", "Directory to scan (defaults to git repository root)")
	scanCmd.Flags().StringVarP(&opts.Out, "out", "o", "index.json", "Path to save the output file")

	return scanCmd
}

func runScan(out io.Writer, opts *ScanOptions) error {
	scanDir := opts.Dir
	if scanDir == "" {
		var err error
		scanDir, err = scanner.FindGitRoot(".")
		if err != nil {
			return err
		}
	}

	fmt.Fprintf(out, "Scanning directory: %s, output: %s\n", scanDir, opts.Out)
	return indexer.BuildIndex(scanDir, opts.Out)
}
