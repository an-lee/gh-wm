package main

import (
	"os"

	"github.com/an-lee/gh-wm/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
