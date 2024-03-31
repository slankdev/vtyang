package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/slankdev/vtyang/pkg/util"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	if err := NewCommand().Execute(); err != nil {
		os.Exit(1)
	}
}

func NewCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:  "frr-agent",
		RunE: f,
	}
	rootCmd.AddCommand(util.NewCommandCompletion(rootCmd))
	rootCmd.AddCommand(util.NewCommandVersion())
	return rootCmd
}

func f(cmd *cobra.Command, args []string) error {
	fmt.Println("NOT_IMPLEMENTED")
	return nil
}
