package command

import (
	"casefile/internal/core"
	"casefile/internal/core/tool"
	"casefile/internal/provider"
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

	p := openai.New(providerConfig)

	req := provider.Request{
		Messages: []provider.Message{{Role: provider.RoleUser, Content: args[0]}},
		Tools:    make([]tool.Tool, 0),
	}

	res, err := p.Complete(context.Background(), req)
	if err != nil {
		return err
	}

	fmt.Println(res.Content)

	return nil
}
