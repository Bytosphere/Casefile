package command

import (
	"casefile/internal/core"
	"fmt"

	"github.com/spf13/cobra"
)

// initCmd is a command that creates the Casefile State (`.casefile/`) directory.
var initCmd = &cobra.Command{
	Use:   "init PATH",
	Short: "Initialize a new State",
	Args:  cobra.ExactArgs(1),
	RunE:  runInit,
}

func runInit(_ *cobra.Command, args []string) error {
	path := args[0]

	// Create the State object.
	state, err := core.NewState(path)

	if err != nil {
		return err
	}

	fmt.Printf("init: state initialized at %s\n", state.AbsolutePath())

	return nil
}
