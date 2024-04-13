package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/k0kubun/pp"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	vtyangapi "github.com/slankdev/vtyang/pkg/grpc/api"
	"github.com/slankdev/vtyang/pkg/linux-agent/yang"
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
	conn, err := grpc.Dial(
		clioptVtyang,
		grpc.WithTimeout(1*time.Second),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return err
	}
	pp.Println("connected")
	defer conn.Close()
	client := vtyangapi.NewGreetingServiceClient(conn)
	stream, err := client.HelloStream(context.Background())
	if err != nil {
		return err
	}
	loop := true
	for loop {
		res, err := stream.Recv()
		if err != nil {
			fmt.Println(err.Error())
			loop = false
			continue
		}
		device := yang.Device{}
		if err := json.Unmarshal([]byte(res.Data), &device); err != nil {
			return err
		}
		pp.Println(device)
		if err := validate(&device); err != nil {
			return err
		}
		if err := commit(&device); err != nil {
			return err
		}
	}
	return nil
}

func validate(_ *yang.Device) error {
	// TODO(slankdev): it's not implemented
	return nil
}

func commit(device *yang.Device) error {
	fmt.Println("NOT IMPLEMENTED (COMMIT)")
	return nil
}
