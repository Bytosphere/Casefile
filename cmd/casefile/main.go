package main

import (
	"casefile/internal/command"
	"os"
)

func main() {
	if err := command.Execute(); err != nil {
		os.Exit(0)
	}
}
