package command

import (
	"github.com/spf13/cobra"
)

// version is the main version of the CLI.
var version = "0.1.0"

// root is the command that all subcommands belong to.
var root = &cobra.Command{
	Use:     "casefile",
	Short:   "Casefile — AI-driven codebase auditing tool",
	Long:    "Casefile scans a codebase for issues, tracks them in a local state, and generates severity-ordered reports.",
	Version: version,
}

// Execute runs the CLI flow.
func Execute() error {
	return root.Execute()
}
