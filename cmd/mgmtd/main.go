package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	//"github.com/k0kubun/pp"

	"github.com/spf13/cobra"

	"github.com/slankdev/vtyang/pkg/mgmtd"
	"github.com/slankdev/vtyang/pkg/util"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	if err := NewCommand().Execute(); err != nil {
		os.Exit(1)
	}
}

var (
	clioptFrr    string
	clioptVtyang string
)

func NewCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:  "frr-agent",
		RunE: f,
	}
	rootCmd.Flags().StringVar(&clioptFrr, "frr",
		"localhost:9001", "")
	rootCmd.Flags().StringVar(&clioptVtyang, "vtyang",
		"192.168.64.1:8080", "")
	rootCmd.AddCommand(util.NewCommandCompletion(rootCmd))
	rootCmd.AddCommand(util.NewCommandVersion())
	return rootCmd
}

func f(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	_ = ctx

	name := "vtyang"

	msg := mgmtd.FeMessage_RegisterReq{}
	msg.RegisterReq = &mgmtd.FeRegisterReq{
		ClientName: &name,
	}

	s := msg.RegisterReq.String()

	fmt.Println("HELLLO")
	fmt.Println(s)
	return nil
}
