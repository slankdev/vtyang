package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"

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
	name         = "vtyang"
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
		Use:  "mgmtd",
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
	// STEP1
	conn, err := net.Dial("unix", "/var/run/frr/mgmtd_fe.sock")
	if err != nil {
		return errors.Wrap(err, "net.Dial")
	}
	defer conn.Close()

	// STEP2
	if err := writeProtoBufMsg(conn, &mgmtd.FeMessage{
		Message: &mgmtd.FeMessage_RegisterReq{
			RegisterReq: &mgmtd.FeRegisterReq{
				ClientName: &name,
			},
		},
	}); err != nil {
		return errors.Wrap(err, "writeProtoBufMsg")
	}

	// STEP3
	if err := writeProtoBufMsg(conn, &mgmtd.FeMessage{
		Message: &mgmtd.FeMessage_SessionReq{
			SessionReq: &mgmtd.FeSessionReq{
				Create: util.NewBoolPointer(true),
				Id: &mgmtd.FeSessionReq_ClientConnId{
					ClientConnId: 0,
				},
			},
		},
	}); err != nil {
		return errors.Wrap(err, "writeProtoBufMsg")
	}

	// STEP4
	msg, err := readProtoBufMsg(conn)
	if err != nil {
		return errors.Wrap(err, "readProtoBufMsg")
	}
	fmt.Println(msg.String())
	sessionId := msg.GetSessionReply().SessionId
	pp.Println(sessionId)

	// STEP5
	dsId := mgmtd.DatastoreId_RUNNING_DS
	xpath := "/"
	if err := writeProtoBufMsg(conn, &mgmtd.FeMessage{
		Message: &mgmtd.FeMessage_GetReq{
			GetReq: &mgmtd.FeGetReq{
				SessionId: sessionId,
				Config:    util.NewBoolPointer(true),
				DsId:      &dsId,
				ReqId:     util.NewUint64Pointer(0),
				Data: []*mgmtd.YangGetDataReq{
					{
						Data: &mgmtd.YangData{
							Xpath: &xpath,
						},
						NextIndx: util.NewInt64Pointer(-1),
					},
				},
			},
		},
	}); err != nil {
		return errors.Wrap(err, "writeProtoBufMsg")
	}
	msg1, err := readProtoBufMsg(conn)
	if err != nil {
		return errors.Wrap(err, "readProtoBufMsg")
	}
	fmt.Println(msg1.GetGetReply().String())

	// STEP6

	// STEP99
	time.Sleep(1000 * time.Second)
	return nil
}

func writeProtoBufMsg(conn net.Conn, msg *mgmtd.FeMessage) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return errors.Wrap(err, "proto.Marshal")
	}

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
	if _, err := conn.Write(buf.Bytes()); err != nil {
		return errors.Wrap(err, "conn.Write")
	}
	fmt.Println(hex.Dump(buf.Bytes()))
	return nil
}

func readProtoBufMsg(conn net.Conn) (*mgmtd.FeMessage, error) {
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, errors.Wrap(err, "conn.Read")
	}
	if n > 1024 {
		return nil, errors.Errorf("msg too big (>1024)")
	}
	buff := bytes.NewBuffer(buf)

	// Parse Marker
	marker := uint32(0)
	if err := binary.Read(buff, binary.LittleEndian, &marker); err != nil {
		return nil, errors.Wrap(err, "binary.Read")
	}
	if marker != MGMT_MSG_MARKER_PROTOBUF {
		return nil, errors.Errorf("not PROTOBUF marker")
	}

	// Parse Size
	msize := uint32(0)
	if err := binary.Read(buff, binary.LittleEndian, &msize); err != nil {
		return nil, errors.Wrap(err, "binary.Read")
	}
	msize0 := msize - 8

	// Parse Body
	body := buf[8 : 8+msize0]
	msg := mgmtd.FeMessage{}
	if err := proto.Unmarshal(body, &msg); err != nil {
		return nil, errors.Wrap(err, "proto.Unmarshal")
	}
	return &msg, nil
}
