package main

import (
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"time"

	"github.com/pkg/errors"
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
	name = "vtyang"
	sock = "/var/run/frr/mgmtd_fe.sock"
)

func NewCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:  "mgmtd",
		RunE: f,
	}
	rootCmd.AddCommand(util.NewCommandCompletion(rootCmd))
	rootCmd.AddCommand(util.NewCommandVersion())
	return rootCmd
}

func f(cmd *cobra.Command, args []string) error {
	// Init Client
	client, err := mgmtd.NewClient(sock, name)
	if err != nil {
		return errors.Wrap(err, "mgmtd.NewClient")
	}
	defer client.Close()

	// STEP1: get config /
	fmt.Println("[+] STEP1 get running-config")
	ret, err := client.GetReq(&mgmtd.FeGetReq{
		SessionId: client.GetSessionId(),
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
	})
	if err != nil {
		return errors.Wrap(err, "mgmtd.GetReq")
	}
	for _, data := range ret {
		fmt.Println(data)
	}

	// STEP2: lock running_ds
	fmt.Println("[+] STEP2 Lock")
	if err := client.LockReq(&mgmtd.FeLockDsReq{
		SessionId: client.GetSessionId(),
		ReqId:     util.NewUint64Pointer(0),
		DsId:      mgmtd.DatastoreId_RUNNING_DS.Enum(),
		Lock:      util.NewBoolPointer(true),
	}); err != nil {
		return errors.Wrap(err, "client.LockReq(running_ds)")
	}
	if err := client.LockReq(&mgmtd.FeLockDsReq{
		SessionId: client.GetSessionId(),
		ReqId:     util.NewUint64Pointer(0),
		DsId:      mgmtd.DatastoreId_CANDIDATE_DS.Enum(),
		Lock:      util.NewBoolPointer(true),
	}); err != nil {
		return errors.Wrap(err, "client.LockReq(candidate_ds)")
	}

	// STEP3
	// mgmt set-config /frr-filter:lib/prefix-list[type='ipv4'][name='hoge']/entry[sequence='10']/action permit
	// mgmt set-config /frr-filter:lib/prefix-list[type='ipv4'][name='hoge']/entry[sequence='10']/ipv4-prefix 10.255.0.0/16
	// mgmt set-config /frr-filter:lib/prefix-list[type='ipv4'][name='hoge']/entry[sequence='10']/ipv4-prefix-length-lesser-or-equal 32
	fmt.Println("[+] STEP3: set-config")
	if err := client.SetConfig(&mgmtd.FeSetConfigReq{
		SessionId:      client.GetSessionId(),
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
	}); err != nil {
		return errors.Wrap(err, "client.SetConfig")
	}

	// STEP4: get config /
	fmt.Println("STEP4 get running-config")
	configRunningDs, err := client.GetReq(&mgmtd.FeGetReq{
		SessionId: client.GetSessionId(),
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
	})
	if err != nil {
		return errors.Wrap(err, "mgmtd.GetReq")
	}
	for _, data := range configRunningDs {
		fmt.Println(data)
	}

	// STEP5: get config /
	fmt.Println("STEP5 get candidate-config")
	configCandidateDs, err := client.GetReq(&mgmtd.FeGetReq{
		SessionId: client.GetSessionId(),
		Config:    util.NewBoolPointer(true),
		DsId:      mgmtd.DatastoreId_CANDIDATE_DS.Enum(),
		ReqId:     util.NewUint64Pointer(0),
		Data: []*mgmtd.YangGetDataReq{
			{
				Data: &mgmtd.YangData{
					Xpath: util.NewStringPointer("/"),
				},
				NextIndx: util.NewInt64Pointer(0),
			},
		},
	})
	if err != nil {
		return errors.Wrap(err, "mgmtd.GetReq")
	}
	for _, data := range configCandidateDs {
		fmt.Println(data)
	}

	// STEP6: commit check & apply
	fmt.Println("STEP6 commit if-diff")
	if !reflect.DeepEqual(configRunningDs, configCandidateDs) {
		fmt.Println("commit")
		if err := client.CommitConfig(&mgmtd.FeCommitConfigReq{
			SessionId:    client.GetSessionId(),
			ReqId:        util.NewUint64Pointer(0),
			SrcDsId:      mgmtd.DatastoreId_CANDIDATE_DS.Enum(),
			DstDsId:      mgmtd.DatastoreId_RUNNING_DS.Enum(),
			ValidateOnly: util.NewBoolPointer(false),
			Abort:        util.NewBoolPointer(false),
		}); err != nil {
			return errors.Wrap(err, "mgmtd.WriteProtoBufMsg")
		}
	}

	// STEP99
	fmt.Println("WAIT 1000s")
	time.Sleep(1000 * time.Second)
	return nil
}
