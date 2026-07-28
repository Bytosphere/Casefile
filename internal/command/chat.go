package command

import (
	"casefile/internal/core"
	"casefile/internal/provider/openai"
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// chatCmd is a test command for communicating with a provider.
var chatCmd = &cobra.Command{
	Use:   "chat MESSAGE",
	Short: "Test talking to a provider",
	Args:  cobra.ExactArgs(1),
	RunE:  runChat,
}

func init() {
	rootCmd.AddCommand(chatCmd)
}

func runChat(_ *cobra.Command, args []string) error {
	state, err := core.LoadState()
	if err != nil {
		return err
	}

	// Grab the configuration for the current profile.
	config := state.Profile().Config()
	providerConfig := config.Provider

	provider := openai.New(providerConfig)

	res, err := provider.Complete(context.Background(), args[0])
	if err != nil {
		return err
	}

	fmt.Println(res)

	return nil
}
