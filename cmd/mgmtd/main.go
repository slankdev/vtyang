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
	fmt.Println("\n\nSTEP4")
	msgs, err := readProtoBufMsg(conn)
	if err != nil {
		return errors.Wrap(err, "readProtoBufMsg")
	}
	msg := msgs[0]
	fmt.Println(msg.String())
	sessionId := msg.GetSessionReply().SessionId
	pp.Println(sessionId)

	// STEP5: get config /
	fmt.Println("\n\nSTEP5")
	if err := writeProtoBufMsg(conn, &mgmtd.FeMessage{
		Message: &mgmtd.FeMessage_GetReq{
			GetReq: &mgmtd.FeGetReq{
				SessionId: sessionId,
				Config:    util.NewBoolPointer(true),
				DsId:      mgmtd.DatastoreId_RUNNING_DS.Enum(),
				ReqId:     util.NewUint64Pointer(0),
				Data: []*mgmtd.YangGetDataReq{
					{
						Data: &mgmtd.YangData{
							Xpath: util.NewStringPointer("/"),
						},
						NextIndx: util.NewInt64Pointer(0),
					},
				},
			},
		},
	}); err != nil {
		return errors.Wrap(err, "writeProtoBufMsg")
	}
	msg1_, err := readProtoBufMsg(conn)
	if err != nil {
		return errors.Wrap(err, "readProtoBufMsg")
	}
	for _, msg := range msg1_ {
		for _, data := range msg.GetGetReply().Data.Data {
			fmt.Println(data)
		}
	}

	// STEP6: lock running_ds
	fmt.Println("\n\nSTEP6")
	if err := writeProtoBufMsg(conn, &mgmtd.FeMessage{
		Message: &mgmtd.FeMessage_LockdsReq{
			LockdsReq: &mgmtd.FeLockDsReq{
				SessionId: sessionId,
				ReqId:     util.NewUint64Pointer(0),
				DsId:      mgmtd.DatastoreId_RUNNING_DS.Enum(),
				Lock:      util.NewBoolPointer(true),
			},
		},
	}); err != nil {
		return errors.Wrap(err, "writeProtoBufMsg")
	}
	msg2_, err := readProtoBufMsg(conn)
	if err != nil {
		return errors.Wrap(err, "readProtoBufMsg")
	}
	for _, msg := range msg2_ {
		fmt.Println(msg.String())
	}

	// STEP7: lock candidate_ds
	fmt.Println("\n\nSTEP7")
	if err := writeProtoBufMsg(conn, &mgmtd.FeMessage{
		Message: &mgmtd.FeMessage_LockdsReq{
			LockdsReq: &mgmtd.FeLockDsReq{
				SessionId: sessionId,
				ReqId:     util.NewUint64Pointer(0),
				DsId:      mgmtd.DatastoreId_CANDIDATE_DS.Enum(),
				Lock:      util.NewBoolPointer(true),
			},
		},
	}); err != nil {
		return errors.Wrap(err, "writeProtoBufMsg")
	}
	msg3_, err := readProtoBufMsg(conn)
	if err != nil {
		return errors.Wrap(err, "readProtoBufMsg")
	}
	for _, msg := range msg3_ {
		fmt.Println(msg.String())
	}

	// STEP8
	// mgmt set-config /frr-filter:lib/prefix-list[type='ipv4'][name='hoge']/entry[sequence='10']/action permit
	// mgmt set-config /frr-filter:lib/prefix-list[type='ipv4'][name='hoge']/entry[sequence='10']/ipv4-prefix 10.255.0.0/16
	// mgmt set-config /frr-filter:lib/prefix-list[type='ipv4'][name='hoge']/entry[sequence='10']/ipv4-prefix-length-lesser-or-equal 32
	fmt.Println("\n\nSTEP8")
	if err := writeProtoBufMsg(conn, &mgmtd.FeMessage{
		Message: &mgmtd.FeMessage_SetcfgReq{
			SetcfgReq: &mgmtd.FeSetConfigReq{
				SessionId:      sessionId,
				DsId:           mgmtd.DatastoreId_CANDIDATE_DS.Enum(),
				CommitDsId:     mgmtd.DatastoreId_RUNNING_DS.Enum(),
				ReqId:          util.NewUint64Pointer(0),
				ImplicitCommit: util.NewBoolPointer(false),
				Data: []*mgmtd.YangCfgDataReq{
					{
						ReqType: mgmtd.CfgDataReqType_SET_DATA.Enum(),
						Data: &mgmtd.YangData{
							Xpath: util.NewStringPointer("/frr-filter:lib/prefix-list[type='ipv4'][name='hoge']/entry[sequence='10']/action"),
							Value: &mgmtd.YangDataValue{
								Value: &mgmtd.YangDataValue_EncodedStrVal{
									EncodedStrVal: "permit",
								},
							},
						},
					},
					{
						ReqType: mgmtd.CfgDataReqType_SET_DATA.Enum(),
						Data: &mgmtd.YangData{
							Xpath: util.NewStringPointer("/frr-filter:lib/prefix-list[type='ipv4'][name='hoge']/entry[sequence='10']/ipv4-prefix"),
							Value: &mgmtd.YangDataValue{
								Value: &mgmtd.YangDataValue_EncodedStrVal{
									EncodedStrVal: "10.255.0.0/16",
								},
							},
						},
					},
					{
						ReqType: mgmtd.CfgDataReqType_SET_DATA.Enum(),
						Data: &mgmtd.YangData{
							Xpath: util.NewStringPointer("/frr-filter:lib/prefix-list[type='ipv4'][name='hoge']/entry[sequence='10']/ipv4-prefix-length-lesser-or-equal"),
							Value: &mgmtd.YangDataValue{
								Value: &mgmtd.YangDataValue_EncodedStrVal{
									EncodedStrVal: "32",
								},
							},
						},
					},
				},
			},
		},
	}); err != nil {
		return errors.Wrap(err, "writeProtoBufMsg")
	}
	msg4_, err := readProtoBufMsg(conn)
	if err != nil {
		return errors.Wrap(err, "readProtoBufMsg")
	}
	for _, msg := range msg4_ {
		fmt.Println(msg.String())
	}

	// STEP9: commit check & apply
	fmt.Println("\n\nSTEP9")
	if err := writeProtoBufMsg(conn, &mgmtd.FeMessage{
		Message: &mgmtd.FeMessage_CommcfgReq{
			CommcfgReq: &mgmtd.FeCommitConfigReq{
				SessionId:    sessionId,
				ReqId:        util.NewUint64Pointer(0),
				SrcDsId:      mgmtd.DatastoreId_CANDIDATE_DS.Enum(),
				DstDsId:      mgmtd.DatastoreId_RUNNING_DS.Enum(),
				ValidateOnly: util.NewBoolPointer(false),
				Abort:        util.NewBoolPointer(false),
			},
		},
	}); err != nil {
		return errors.Wrap(err, "writeProtoBufMsg")
	}
	msg5_, err := readProtoBufMsg(conn)
	if err != nil {
		return errors.Wrap(err, "readProtoBufMsg")
	}
	for _, msg := range msg5_ {
		fmt.Println(msg.String())
	}

	// STEP99
	fmt.Println("WAIT 1000s")
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

func readProtoBufMsg(conn net.Conn) ([]*mgmtd.FeMessage, error) {
	const size = 40960
	buf := make([]byte, size)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, errors.Wrap(err, "conn.Read")
	}
	if n >= size {
		return nil, errors.Errorf("msg too big (>%d)", size)
	}
	buff := bytes.NewBuffer(buf)
	msgs := []*mgmtd.FeMessage{}

	remain := n
	for remain > 0 {
		// Parse Marker
		marker := uint32(0)
		if err := binary.Read(buff, binary.LittleEndian, &marker); err != nil {
			return nil, errors.Wrap(err, "binary.Read(marker)")
		}
		if marker != MGMT_MSG_MARKER_PROTOBUF {
			return nil, errors.Errorf("not PROTOBUF marker")
		}

		// Parse Size
		msize := uint32(0)
		if err := binary.Read(buff, binary.LittleEndian, &msize); err != nil {
			return nil, errors.Wrap(err, "binary.Read(size)")
		}
		msize0 := msize - 8
		remain -= int(msize)

		// Parse Body
		body := make([]byte, msize0)
		if err := binary.Read(buff, binary.LittleEndian, &body); err != nil {
			pp.Println(buf)
			return nil, errors.Wrap(err, "binary.Read(body)")
		}
		msg := &mgmtd.FeMessage{}
		if err := proto.Unmarshal(body, msg); err != nil {
			return nil, errors.Wrap(err, "proto.Unmarshal")
		}
		msgs = append(msgs, msg)
	}
	return msgs, nil
}
