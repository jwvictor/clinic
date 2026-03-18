package main

import (
	"os"

	"github.com/togglemedia/cliq/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
