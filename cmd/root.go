package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	opts := &ScanOptions{}
	
	rootCmd := &cobra.Command{
		Use:   "hydrogen-index",
		Short: "A tool to index Hydrogen artifacts for online import",
		Run: func(cmd *cobra.Command, args []string) {
			// Default to scan
			runScan(cmd.OutOrStdout(), opts)
		},
	}
	
	// Add flags to root for the default 'scan' behavior
	rootCmd.Flags().StringVarP(&opts.Dir, "dir", "d", "", "Directory to scan (defaults to git repository root)")
	rootCmd.Flags().StringVarP(&opts.Out, "out", "o", "index.json", "Path to save the output file")

	rootCmd.AddCommand(NewScanCmd())
	rootCmd.AddCommand(NewValidateCmd())
	rootCmd.AddCommand(NewVersionCmd())

	return rootCmd
}

func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
