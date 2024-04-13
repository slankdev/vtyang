package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/k0kubun/pp"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/slankdev/vtyang/pkg/frr"
	frrapi "github.com/slankdev/vtyang/pkg/frr"
	vtyangapi "github.com/slankdev/vtyang/pkg/grpc/api"
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

	// Init VTYANG gRPC Client
	conn, err := grpc.DialContext(
		ctx,
		clioptVtyang,
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

	// Init FRR gRPC Client
	connFrr, err := grpc.DialContext(
		ctx,
		clioptFrr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return err
	}
	pp.Println("frr connected")
	defer connFrr.Close()
	clientFrr := frrapi.NewNorthboundClient(connFrr)

	// Main loop
	loop := true
	for loop {
		res, err := stream.Recv()
		if err != nil {
			fmt.Println(err.Error())
			loop = false
			continue
		}
		if err := commit(res.Data, clientFrr); err != nil {
			return err
		}
	}
	return nil
}

func commit(data string, clientFrr frr.NorthboundClient) error {
	// Create Candidate
	ctx := context.Background()
	resp, err := clientFrr.CreateCandidate(ctx, &frrapi.CreateCandidateRequest{})
	if err != nil {
		return err
	}
	pp.Println(resp.CandidateId)

	// Load config to candidate
	fmt.Println(data)

	// Commit
	return nil
}
