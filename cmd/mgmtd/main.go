package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"

	//"github.com/k0kubun/pp"

	"github.com/k0kubun/pp"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

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

const (
	MGMT_MSG_MARKER_PFX      = uint32(0x23232300)
	MGMT_MSG_MARKER_PROTOBUF = uint32(MGMT_MSG_MARKER_PFX | 0x0)
	MGMT_MSG_MARKER_NATIVE   = uint32(MGMT_MSG_MARKER_PFX | 0x1)
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

	// STEP1
	conn, err := net.Dial("unix", "/var/run/frr/mgmtd_fe.sock")
	if err != nil {
		return errors.Wrap(err, "net.Dial")
	}
	defer conn.Close()

	// STEP2
	msg := mgmtd.FeMessage{
		Message: &mgmtd.FeMessage_RegisterReq{
			RegisterReq: &mgmtd.FeRegisterReq{
				ClientName: &name,
			},
		},
	}
	data, err := proto.Marshal(&msg)
	if err != nil {
		return errors.Wrap(err, "proto.Marshal")
	}
	pp.Println(data)

	// STEP3
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.NativeEndian,
		MGMT_MSG_MARKER_PROTOBUF); err != nil {
		return errors.Wrap(err, "binary.Write")
	}
	if err := binary.Write(buf, binary.NativeEndian,
		uint32(8+len(data))); err != nil {
		return errors.Wrap(err, "binary.Write")
	}
	if _, err := buf.Write([]byte(data)); err != nil {
		return errors.Wrap(err, "buf.Write")
	}
	n, err := conn.Write(buf.Bytes())
	if err != nil {
		return errors.Wrap(err, "conn.Write")
	}
	pp.Println(n)

	out := hex.Dump(buf.Bytes())
	fmt.Println(out)
	time.Sleep(1000 * time.Second)
	return nil
}
