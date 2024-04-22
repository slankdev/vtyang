package mgmtd

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"

	"github.com/k0kubun/pp"
	"github.com/pkg/errors"
	"github.com/slankdev/vtyang/pkg/util"
	"google.golang.org/protobuf/proto"
)

const (
	MGMT_MSG_MARKER_PFX      = uint32(0x23232300)
	MGMT_MSG_MARKER_PROTOBUF = uint32(0x23232300)
	MGMT_MSG_MARKER_NATIVE   = uint32(0x23232301)
)

var (
	globalVerboseEnabled = false
)

type Client struct {
	conn      net.Conn
	sessionId uint64
}

func SetGlobalVerbose(enabled bool) {
	globalVerboseEnabled = enabled
}

func NewClient(sockpath, name string) (*Client, error) {
	// Connect to socket
	conn, err := net.Dial("unix", "/var/run/frr/mgmtd_fe.sock")
	if err != nil {
		return nil, errors.Wrap(err, "net.Dial")
	}

	// Register FE-Client
	if err := writeProtobufMsg(conn, &FeMessage{
		Message: &FeMessage_RegisterReq{
			RegisterReq: &FeRegisterReq{
				ClientName: &name,
			},
		},
	}); err != nil {
		return nil, errors.Wrap(err, "writeProtobufMsg(RegisterReq)")
	}

	// Create FE-Session
	if err := writeProtobufMsg(conn, &FeMessage{
		Message: &FeMessage_SessionReq{
			SessionReq: &FeSessionReq{
				Create: util.NewBoolPointer(true),
				Id: &FeSessionReq_ClientConnId{
					ClientConnId: 0,
				},
			},
		},
	}); err != nil {
		return nil, errors.Wrap(err, "writeProtobufMsg(SessionReq)")
	}
	msgs, err := readProtobufMsg(conn)
	if err != nil {
		return nil, errors.Wrap(err, "readProtobufMsg(SessionReq)")
	}
	msg := msgs[0]
	sessionId := msg.GetSessionReply().SessionId

	// Return client data
	return &Client{
		conn:      conn,
		sessionId: *sessionId,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) GetRaw() (net.Conn, uint64) {
	return c.conn, c.sessionId
}

func (c *Client) GetSessionId() *uint64 {
	return &c.sessionId
}

func (c *Client) GetReq(req *FeGetReq) ([]*YangData, error) {
	if err := writeProtobufMsg(c.conn, &FeMessage{
		Message: &FeMessage_GetReq{GetReq: req}}); err != nil {
		return nil, errors.Wrap(err, "writeProtobufMsg")
	}
	msgs, err := readProtobufMsg(c.conn)
	if err != nil {
		return nil, errors.Wrap(err, "readProtobufMsg")
	}
	ret := []*YangData{}
	for _, msg := range msgs {
		ret = append(ret, msg.GetGetReply().Data.Data...)
	}
	return ret, nil
}

func (c *Client) LockReq(req *FeLockDsReq) error {
	if err := writeProtobufMsg(c.conn, &FeMessage{
		Message: &FeMessage_LockdsReq{LockdsReq: req}}); err != nil {
		return errors.Wrap(err, "writeProtobufMsg")
	}
	msgs, err := readProtobufMsg(c.conn)
	if err != nil {
		return errors.Wrap(err, "readProtobufMsg")
	}
	for _, msg := range msgs {
		if errIfAny := msg.GetLockdsReply().ErrorIfAny; errIfAny != nil {
			return errors.Errorf("LockReq(reply-error): %s", *errIfAny)
		}
	}
	return nil
}

func (c *Client) SetConfig(req *FeSetConfigReq) error {
	if err := writeProtobufMsg(c.conn, &FeMessage{
		Message: &FeMessage_SetcfgReq{SetcfgReq: req}}); err != nil {
		return errors.Wrap(err, "writeProtobufMsg")
	}
	msgs, err := readProtobufMsg(c.conn)
	if err != nil {
		return errors.Wrap(err, "readProtobufMsg")
	}
	for _, msg := range msgs {
		if errIfAny := msg.GetSetcfgReply().ErrorIfAny; errIfAny != nil {
			return errors.Errorf("SetConfig(reply-error): %s", *errIfAny)
		}
	}
	return nil
}

func (c *Client) CommitConfig(req *FeCommitConfigReq) error {
	if err := writeProtobufMsg(c.conn, &FeMessage{
		Message: &FeMessage_CommcfgReq{CommcfgReq: req}}); err != nil {
		return errors.Wrap(err, "writeProtobufMsg")
	}
	msgs, err := readProtobufMsg(c.conn)
	if err != nil {
		return errors.Wrap(err, "readProtobufMsg")
	}
	for _, msg := range msgs {
		if errIfAny := msg.GetCommcfgReply().ErrorIfAny; errIfAny != nil {
			return errors.Errorf("CommitConfig(reply-error): %s", *errIfAny)
		}
	}
	return nil
}

func writeProtobufMsg(conn net.Conn, msg *FeMessage) error {
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
	if globalVerboseEnabled {
		fmt.Println(hex.Dump(buf.Bytes()))
	}
	return nil
}

func readProtobufMsg(conn net.Conn) ([]*FeMessage, error) {
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
	msgs := []*FeMessage{}

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
		msg := &FeMessage{}
		if err := proto.Unmarshal(body, msg); err != nil {
			return nil, errors.Wrap(err, "proto.Unmarshal")
		}
		msgs = append(msgs, msg)
	}
	return msgs, nil
}
