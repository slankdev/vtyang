package vtyang

import "github.com/spf13/cobra"

func agentMain(cmd *cobra.Command, args []string) error {
	print("hello\n")
	return nil
}
