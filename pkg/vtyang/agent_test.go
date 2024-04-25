package vtyang

import (
	"fmt"
	"os"
	"testing"

	"github.com/slankdev/vtyang/pkg/util"
)

const (
	RUNTIME_PATH = "/tmp/run/vtyang" // TODO
)

func TestAgentNoDatabase(t *testing.T) {
	executeTestCase(t, &TestCase{
		YangPath:    "./testdata",
		RuntimePath: "/tmp/run/vtyang",
		OutputFile:  "./testdata/no_database/output.txt",
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
		InitConfigFile: "./testdata/load_database/config.json",
		YangPath:       "./testdata",
		RuntimePath:    "/tmp/run/vtyang",
		OutputFile:     "./testdata/load_database/output.txt",
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
		YangPath:    "../../yang.frr/",
		RuntimePath: "/tmp/run/vtyang",
		OutputFile:  "./testdata/xpath_cli/output.txt",
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
		YangPath:    "./testdata/frr_isid_test1",
		RuntimePath: "/tmp/run/vtyang",
		OutputFile:  "./testdata/frr_isid_test1/output.json",
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
		YangPath:    "./testdata/same_container_name_in_different_modules/",
		RuntimePath: "/tmp/run/vtyang",
		OutputFile:  "./testdata/same_container_name_in_different_modules/clitree.json",
		Inputs: []string{
			"show cli-tree",
		},
	})
}

func TestLoadDatabaseFromFile(t *testing.T) {
	executeTestCase(t, &TestCase{
		YangPath:       "../../yang.frr/",
		RuntimePath:    "/tmp/run/vtyang",
		InitConfigFile: "./testdata/runtime1/config.json",
		OutputFile:     "./testdata/runtime1/output.txt",
		Inputs: []string{
			"show running-config-frr",
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
							Type:    YInteger,
							Integer: 65001,
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
							Type:    YInteger,
							Integer: 65001,
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
											Type:   YString,
											String: "1",
										},
									},
									{
										Name: "vrf",
										Type: Leaf,
										Value: DBValue{
											Type:   YString,
											String: "default",
										},
									},
									{
										Name: "area-address",
										Type: LeafList,
										Value: DBValue{
											Type: YStringArray,
											StringArray: []string{
												"10.0000.0000.0000.0000.0000.0000.0000.0000.0000.00",
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
											Type:   YString,
											String: "1",
										},
									},
									{
										Name: "vrf",
										Type: Leaf,
										Value: DBValue{
											Type:   YString,
											String: "default",
										},
									},
									{
										Name: "area-address",
										Type: LeafList,
										Value: DBValue{
											Type: YStringArray,
											StringArray: []string{
												"10.0000.0000.0000.0000.0000.0000.0000.0000.0000.00",
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
		"../../yang.frr/", "/tmp/testlog.log"); err != nil {
		t.Fatal(err)
	}

	result, err := filterDbWithModule(input, "frr-isisd")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(result.String())

	diff := DBNodeDiff(result, expected)
	if diff != "" {
		t.Fatalf("diff %s\n", diff)
	}
}
