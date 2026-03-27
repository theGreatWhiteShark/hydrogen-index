package main

import (
	"os"

	"github.com/hydrogen-music/hydrogen-index/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
