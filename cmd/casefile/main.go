package main

import "casefile/internal/command"

func main() {
	if err := command.Execute(); err != nil {
		panic(err)
	}
}
