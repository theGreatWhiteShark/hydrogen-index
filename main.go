package main

import (
	"fmt"
	"os"

	"github.com/theGreatWhiteShark/hydrogen-index/cmd"
)

const version = "0.1.0"

func main() {
	command := cmd.NewRootCommand(cmd.Dependencies{
		WorkingDir: mustGetwd(),
		Stdout:     os.Stdout,
		Stderr:     os.Stderr,
		Version:    version,
	})

	if err := command.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func mustGetwd() string {
	workingDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return workingDir
}
