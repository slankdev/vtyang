package vtyang

import (
	"os"
	"testing"

	"github.com/openconfig/goyang/pkg/yang"
	"github.com/slankdev/vtyang/pkg/util"
)

const (
	RUNTIME_PATH = "/tmp/run/vtyang" // TODO
)

func TestAgentNoDatabase(t *testing.T) {
	executeTestCase(t, &TestCase{
		RuntimePath: "/tmp/run/vtyang",
		YangPath:    "./testdata/yang/accounting",
		OutputFile:  "./testdata/no_database_output.txt",
		Inputs: []string{
			"show running-config",
			"configure",
			"set users user hiroki",
			"commit",
			"do show running-config",
		},
	})
}

func TestAgentLoadDatabase(t *testing.T) {
	executeTestCase(t, &TestCase{
		RuntimePath:    "/tmp/run/vtyang",
		YangPath:       "./testdata/yang/accounting",
		InitConfigFile: "./testdata/load_database_config.json",
		OutputFile:     "./testdata/load_database_output.txt",
		Inputs: []string{
			"show running-config",
			"configure",
			"set users user shirokura projects mfplane",
			"set users user shirokura age 28",
			"set users user hiroki age 23",
			"commit",
			"do show running-config",
		},
	})
}

func TestAgentXPathCli(t *testing.T) {
	executeTestCase(t, &TestCase{
		RuntimePath: "/tmp/run/vtyang",
		YangPath:    "./testdata/yang/frr_isisd_minimal",
		OutputFile:  "./testdata/xpath_cli_output.txt",
		Inputs: []string{
			"configure",
			"set isis instance 1 default description area1-default-hoge",
			"set isis instance 1 vrf0 description area1-vrf0-hoge",
			"set isis instance 2 vrf0 description area2-vrf0-hoge",
			"set isis instance 1 vrf0 description area1-vrf0-fuga",
			"set isis instance 1 default area-address 10.0000.0000.0000.0000.0000.0000.0000.0000.0000.00 20.0000.0000.0000.0000.0000.0000.0000.0000.0000.00",
			"commit",
			"do show running-config",
		},
	})
}

func TestAgentXPathCliFRR(t *testing.T) {
	executeTestCase(t, &TestCase{
		RuntimePath: "/tmp/run/vtyang",
		LogFile:     "/tmp/run/vtyang/vtyang.log",
		YangPath:    "./testdata/yang/frr_isisd_minimal",
		OutputFile:  "./testdata/frr_isisd_test1_output.json",
		Inputs: []string{
			"configure",
			"set isis instance 1 default area-address 10.0000.0000.0000.0000.0000.0000.0000.0000.0000.00",
			"set isis instance 1 default flex-algos flex-algo 128 priority 100",
			"commit",
			"do show running-config-frr",
		},
	})
}

func TestYangCompletion1(t *testing.T) {
	executeTestCase(t, &TestCase{
		RuntimePath: "/tmp/run/vtyang",
		YangPath:    "./testdata/yang/same_container_name_in_different_modules/",
		OutputFile:  "./testdata/same_container_name_in_different_modules_output.json",
		Inputs: []string{
			"show cli-tree",
		},
	})
}

func TestLoadDatabaseFromFile(t *testing.T) {
	executeTestCase(t, &TestCase{
		RuntimePath:    "/tmp/run/vtyang",
		YangPath:       "./testdata/yang/frr_isisd_minimal",
		InitConfigFile: "./testdata/show_run_frr_config.json",
		OutputFile:     "./testdata/show_run_frr_output.txt",
		Inputs: []string{
			"show running-config-frr",
		},
	})
}

func TestXpathParse1(t *testing.T) {
	executeTestCase(t, &TestCase{
		RuntimePath: "/tmp/run/vtyang",
		YangPath:    "./testdata/yang/frr_mgmtd_minimal",
		OutputFile:  "./testdata/xpath_parse1_output.txt",
		Inputs: []string{
			"show-xpath lib interface dum0 description dum0-comment",
		},
	})
}

func TestXpathParse2(t *testing.T) {
	executeTestCase(t, &TestCase{
		RuntimePath: "/tmp/run/vtyang",
		LogFile:     "/tmp/run/vtyang/vtyang.log",
		YangPath:    "./testdata/yang/leaf_types",
		OutputFile:  "./testdata/xpath_parse2_output.txt",
		Inputs: []string{
			"configure",
			"set values u08 0",
			"set values u16 0",
			"set values u32 0",
			"set values u64 0",
			"set values i08 -128",
			"set values i16 -32768",
			"set values i32 -2147483648",
			"set values i64 -9223372036854775808",
			"set values percentage 0",
			"set values month 1",
			"set values decimal -0.22",
			"set values bool false",
			"set values crypto des3",
			"commit",
			"quit",
			"show running-config",

			"configure",
			"set values u08 255",
			"set values u16 65535",
			"set values u32 4294967295",
			"set values u64 18446744073709551615",
			"set values i08 127",
			"set values i16 32767",
			"set values i32 2147483647",
			"set values i64 9223372036854775807",
			"set values percentage 100",
			"set values month 12",
			"set values decimal 3.14",
			"set values bool true",
			"set values crypto aes",
			"commit",
			"quit",
			"show running-config",

			"configure",
			"delete values",
			"set values items item1 hiroki description hello1",
			"commit",
			"quit",
			"show running-config",

			"configure",
			"delete values",
			"set values items item2 hiroki staticd description hello2",
			"commit",
			"quit",
			"show running-config",

			"configure",
			"delete values",
			"set values items item3 hiroki staticd vrf0 description hello3",
			"commit",
			"quit",
			"show running-config",
		},
	})
}

func TestXpathParse3(t *testing.T) {
	executeTestCase(t, &TestCase{
		RuntimePath: "/tmp/run/vtyang",
		LogFile:     "/tmp/run/vtyang/vtyang.log",
		YangPath:    "./testdata/yang/frr_mgmtd_minimal",
		OutputFile:  "./testdata/xpath_parse3_output.txt",
		Inputs: []string{
			"configure",
			"set lib prefix-list ipv4 hoge entry 10 action permit",
			//"set lib prefix-list ipv4 hoge entry 10 value ipv4-prefix ipv4-prefix 10.255.0.0/16",
			//"set lib prefix-list ipv4 hoge entry 10 ipv4-prefix 10.255.0.0/16",
			// "set lib prefix-list ipv4 hoge entry 10 ipv4-prefix-length-lesser-or-equal 32",
			// "set lib prefix-list ipv4 hoge entry 20 ipv4-prefix 10.254.0.0/16",
			// "set lib prefix-list ipv4 hoge entry 20 ipv4-prefix-length-lesser-or-equal 32",
			// "show configuration diff",
		},
	})
}

func TestXpathParse4(t *testing.T) {
	executeTestCase(t, &TestCase{
		RuntimePath: "/tmp/run/vtyang",
		LogFile:     "/tmp/run/vtyang/vtyang.log",
		YangPath:    "./testdata/yang/leaf_types",
		OutputFile:  "./testdata/xpath_parse4_output.txt",
		Inputs: []string{
			"show-xpath values u08 100",
			"show-xpath values u16 100",
			"show-xpath values u32 100",
			"show-xpath values u64 100",
			"show-xpath values i08 100",
			"show-xpath values i16 100",
			"show-xpath values i32 100",
			"show-xpath values i64 100",
			"show-xpath values month 12",
		},
	})
}

func TestChoiceCase1(t *testing.T) {
	executeTestCase(t, &TestCase{
		RuntimePath: "/tmp/run/vtyang",
		LogFile:     "/tmp/run/vtyang/vtyang.log",
		YangPath:    "./testdata/yang/choice_case",
		OutputFile:  "./testdata/choice_case_output1.json",
		Inputs: []string{
			"show cli-tree",
		},
	})
}

func TestChoiceCase2(t *testing.T) {
	executeTestCase(t, &TestCase{
		RuntimePath: "/tmp/run/vtyang",
		LogFile:     "/tmp/run/vtyang/vtyang.log",
		YangPath:    "./testdata/yang/choice_case",
		OutputFile:  "./testdata/choice_case_output2.txt",
		Inputs: []string{
			"show-xpath values transport-proto tcp-app http",
			"show-xpath items items hoge ipv4-proto icmp",
			"configure",
			"set values transport-proto udp-app dns",
			"set items items icmp4 ipv4-proto icmp",
			"set items items icmp6 ipv6-proto icmp",
			"commit",
			"quit",
			"show running-config",
		},
	})
}

func TestFilterDbWithModule(t *testing.T) {
	input := &DBNode{
		Name: "",
		Type: Container,
		Childs: []DBNode{
			{
				Name: "bgp",
				Type: Container,
				Childs: []DBNode{
					{
						Name: "as-number",
						Type: Leaf,
						Value: DBValue{
							Type:  yang.Yint32,
							Int32: 65001,
						},
					},
				},
			},
			{
				Name: "isis",
				Type: Container,
				Childs: []DBNode{
					{
						Name: "ignored",
						Type: Leaf,
						Value: DBValue{
							Type:  yang.Yint32,
							Int32: 65001,
						},
					},
					{
						Name: "instance",
						Type: List,
						Childs: []DBNode{
							{
								Name: "",
								Type: Container,
								Childs: []DBNode{
									{
										Name: "area-tag",
										Type: Leaf,
										Value: DBValue{
											Type:   yang.Ystring,
											String: "1",
										},
									},
									{
										Name: "vrf",
										Type: Leaf,
										Value: DBValue{
											Type:   yang.Ystring,
											String: "default",
										},
									},
									{
										Name: "area-address",
										Type: LeafList,
										ArrayValue: []DBValue{
											{
												Type:   yang.Ystring,
												String: "10.0000.0000.0000.0000.0000.0000.0000.0000.0000.00",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	expected := &DBNode{
		Name: "",
		Type: Container,
		Childs: []DBNode{
			{
				Name: "frr-isisd:isis",
				Type: Container,
				Childs: []DBNode{
					{
						Name: "instance",
						Type: List,
						Childs: []DBNode{
							{
								Name: "",
								Type: Container,
								Childs: []DBNode{
									{
										Name: "area-tag",
										Type: Leaf,
										Value: DBValue{
											Type:   yang.Ystring,
											String: "1",
										},
									},
									{
										Name: "vrf",
										Type: Leaf,
										Value: DBValue{
											Type:   yang.Ystring,
											String: "default",
										},
									},
									{
										Name: "area-address",
										Type: LeafList,
										ArrayValue: []DBValue{
											{
												Type:   yang.Ystring,
												String: "10.0000.0000.0000.0000.0000.0000.0000.0000.0000.00",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Preparation
	GlobalOptRunFilePath = RUNTIME_PATH
	if util.FileExists(getDatabasePath()) {
		if err := os.Remove(getDatabasePath()); err != nil {
			t.Error(err)
		}
	}

	// Initializing Agent
	if err := InitAgent(RUNTIME_PATH,
		"./testdata/yang/frr_isisd_minimal",
		"/tmp/testlog.log"); err != nil {
		t.Fatal(err)
	}

	result, err := filterDbWithModule(input, "frr-isisd")
	if err != nil {
		t.Fatal(err)
	}
	diff := DBNodeDiff(result, expected)
	if diff != "" {
		t.Fatalf("diff %s\n", diff)
	}
}
