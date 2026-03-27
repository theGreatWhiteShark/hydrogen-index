package cmd

import (
	"fmt"
	"os"

	"github.com/hydrogen-music/hydrogen-index/internal/validator"
	"github.com/spf13/cobra"
)

func buildValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate <file>",
		Short: "Validate a JSON file against the index schema",
		Args:  cobra.ExactArgs(1),
		RunE:  runValidate,
	}
}

func runValidate(cmd *cobra.Command, args []string) error {
	if err := validator.Validate(args[0]); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%s is valid\n", args[0])
	return nil
}
