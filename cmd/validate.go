package cmd

import (
	"fmt"
	"os"

	"github.com/hydrogen-music/hydrogen-index/pkg/validator"
	"github.com/spf13/cobra"
)

func NewValidateCmd() *cobra.Command {
	validateCmd := &cobra.Command{
		Use:   "validate [file]",
		Short: "Check whether a given file matches the expected index.json schema",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "Validating %s...\n", args[0])
			if err := validator.ValidateIndexFile(args[0]); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "File is valid!\n")
		},
	}
	return validateCmd
}
