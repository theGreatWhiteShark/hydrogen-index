package cmd

import (
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	"github.com/theGreatWhiteShark/hydrogen-index/internal/scan"
	internalvalidate "github.com/theGreatWhiteShark/hydrogen-index/internal/validate"
)

type Dependencies struct {
	WorkingDir string
	Stdout     io.Writer
	Stderr     io.Writer
	Version    string
	Now        func() time.Time
}

func NewRootCommand(deps Dependencies) *cobra.Command {
	if deps.Stdout == nil {
		deps.Stdout = io.Discard
	}
	if deps.Stderr == nil {
		deps.Stderr = io.Discard
	}
	if deps.Now == nil {
		deps.Now = time.Now
	}

	var baseURL string
	command := &cobra.Command{
		Use:           "hydrogen-index",
		Short:         "Build and validate Hydrogen artifact indexes",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       deps.Version,
		RunE: func(_ *cobra.Command, _ []string) error {
			return scan.Run(scan.Options{
				WorkingDir: deps.WorkingDir,
				BaseURL:    baseURL,
				Version:    deps.Version,
				Now:        deps.Now,
			})
		},
	}

	command.SetVersionTemplate("{{.Version}}\n")
	command.SetOut(deps.Stdout)
	command.SetErr(deps.Stderr)
	command.PersistentFlags().StringVar(&baseURL, "base-url", "", "Base URL used to derive artifact permalinks")
	command.AddCommand(newScanCommand(deps, &baseURL))
	command.AddCommand(newValidateCommand(deps))

	return command
}

func newScanCommand(deps Dependencies, baseURL *string) *cobra.Command {
	var directory string
	var outputPath string

	command := &cobra.Command{
		Use:   "scan",
		Short: "Scan Hydrogen artifacts and write index.json",
		RunE: func(_ *cobra.Command, _ []string) error {
			return scan.Run(scan.Options{
				WorkingDir: deps.WorkingDir,
				Directory:  directory,
				OutputPath: outputPath,
				BaseURL:    *baseURL,
				Version:    deps.Version,
				Now:        deps.Now,
			})
		},
	}

	command.Flags().StringVarP(&directory, "directory", "d", "", "Directory to scan instead of searching for the git repository root")
	command.Flags().StringVarP(&outputPath, "output", "o", "", "Path of the generated index.json file")

	return command
}

func newValidateCommand(deps Dependencies) *cobra.Command {
	command := &cobra.Command{
		Use:   "validate <index.json>",
		Short: "Validate an index.json file against the bundled schema",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := internalvalidate.ValidateFile(args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintln(deps.Stdout, "valid")
			return nil
		},
	}

	return command
}
