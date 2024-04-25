package vtyang

import (
	"os"
	"testing"

	"github.com/slankdev/vtyang/pkg/util"
)

const (
	RUNTIME_PATH = "/tmp/run/vtyang" // TODO
)

type TestCaseForTestAgent struct { // TODO
	Inputs []string
	Output string
}

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

const TestAgentNoDatabaseOutput3 = `{
  "isis": {
    "instance": [
      {
        "area-address": [
          "10.0000.0000.0000.0000.0000.0000.0000.0000.0000.00",
          "20.0000.0000.0000.0000.0000.0000.0000.0000.0000.00"
        ],
        "area-tag": "1",
        "description": "area1-default-hoge",
        "vrf": "default"
      },
      {
        "area-tag": "1",
        "description": "area1-vrf0-fuga",
        "vrf": "vrf0"
      },
      {
        "area-tag": "2",
        "description": "area2-vrf0-hoge",
        "vrf": "vrf0"
      }
    ]
  }
}
`

func TestAgentXPathCli(t *testing.T) {
	testcases := []TestCaseForTestAgent{
		{
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
			Output: TestAgentNoDatabaseOutput3,
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
		"../../yang.frr/"); err != nil {
		t.Fatal(err)
	}

	// EXECUTE TEST CASES
	for idx, tc := range testcases {
		buf := setStdoutWithBuffer()
		for _, input := range tc.Inputs {
			t.Logf("Testcase[%d] executing %s", idx, input)
			getCommandNodeCurrent().executeCommand(input)
		}
		result := buf.String()
		if tc.Output != result {
			t.Errorf("Unexpected output")
			for _, input := range tc.Inputs {
				t.Errorf("input %+v", input)
			}
			t.Errorf("expect(len=%d) %+v", len(tc.Output), tc.Output)
			t.Errorf("result(len=%d) %+v", len(result), result)
			t.Fatal("quiting test with FAILED result")
		}
		t.Logf("Testcase[%d] output check is succeeded", idx)
	}
}

func TestAgentXPathCliFRR(t *testing.T) {
	executeTestCase(t, &TestCase{
		YangPath:       "./testdata/frr_isid_test1",
		RuntimePath:    "/tmp/run/vtyang",
		OutputJsonFile: "./testdata/frr_isid_test1/output.json",
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
		YangPath:       "./testdata/same_container_name_in_different_modules/",
		RuntimePath:    "/tmp/run/vtyang",
		OutputJsonFile: "./testdata/same_container_name_in_different_modules/clitree.json",
		Inputs: []string{
			"show cli-tree",
		},
	})
}
