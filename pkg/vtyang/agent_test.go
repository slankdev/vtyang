package vtyang

import (
	"fmt"
	"os"
	"testing"

	"github.com/slankdev/vtyang/pkg/util"
)

const (
	YANG_PATH    = "./testdata"
	RUNTIME_PATH = "/tmp/run/vtyang"
)

type TestCaseForTestAgent struct {
	Inputs []string
	Output string
}

type TestCase struct {
	YangPath       string
	RuntimePath    string
	Inputs         []string
	OutputString   string
	OutputJsonFile string
}

func executeTestCase(t *testing.T, tc *TestCase) {
	// Preparation
	GlobalOptRunFilePath = tc.RuntimePath
	if util.FileExists(getDatabasePath()) {
		if err := os.Remove(getDatabasePath()); err != nil {
			t.Fatal(err)
		}
	}

	// Initializing Agent
	if err := InitAgent(tc.RuntimePath, tc.YangPath); err != nil {
		t.Fatal(err)
	}

	out, err := os.ReadFile(tc.OutputJsonFile)
	if err != nil {
		t.Fatal(err)
	}

	// Execute Test commands
	buf := setStdoutWithBuffer()
	for idx, input := range tc.Inputs {
		t.Logf("Testcase[%d] executing %s", idx, input)
		getCommandNodeCurrent().executeCommand(input)
	}
	result := buf.String()
	eq, err := util.DeepEqualJSON(result, string(out))
	if err != nil {
		t.Fatal(err)
	}
	if !eq {
		t.Errorf("Unexpected output")
		for _, input := range tc.Inputs {
			t.Errorf("input %+v", input)
		}
		p := fmt.Sprintf("/tmp/test_fail_output_%s", util.MakeRandomStr(10))
		if err := os.WriteFile(p+"_expected.json", out, os.ModePerm); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p+"_result.json", []byte(result), os.ModePerm); err != nil {
			t.Fatal(err)
		}
		t.Errorf("expect(len=%d) %s_expected.json", len(string(out)), p)
		t.Errorf("result(len=%d) %s_result.json", len(result), p)
		t.Errorf("KINDLY_CLI diff -u %s_expected.json %s_result.json", p, p)
		t.Fatal("quiting test with FAILED result")
	}
	t.Logf("Testcase output check is succeeded")
}

func TestAgentNoDatabase(t *testing.T) {
	testcases := []TestCaseForTestAgent{
		{
			Inputs: []string{
				"show running-config",
			},
			Output: "{}\n",
		},
		{
			Inputs: []string{
				"configure",
				"set users user hiroki",
				"commit",
				"do show running-config",
			},
			Output: TestAgentNoDatabaseOutput2,
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
	if err := InitAgent(RUNTIME_PATH, YANG_PATH); err != nil {
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

func TestAgentLoadDatabase(t *testing.T) {
	testcases := []TestCaseForTestAgent{
		{
			Inputs: []string{
				"show running-config",
			},
			Output: TestAgentLoadDatabaseOutput1,
		},
		{
			Inputs: []string{
				"configure",
				"set users user shirokura projects mfplane",
				"set users user shirokura age 28",
				"commit",
				"do show running-config",
			},
			Output: TestAgentLoadDatabaseOutput2,
		},
		// (3) Delete database node
		// inputs:
		// - configure
		// - set segment-routing ...
		// output: xxx
		// (4) Update database node
		// (5) CLI Completion
	}

	// Preparation
	if err := os.WriteFile(getDatabasePath(), []byte(dbContent), 0644); err != nil {
		t.Error(err)
	}

	// Initializing Agent
	if err := InitAgent(RUNTIME_PATH, YANG_PATH); err != nil {
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

const dbContent = `
{
  "users": {
    "user": [
      {
        "age": 22,
        "name": "hiroki"
      },
      {
        "age": 30,
        "name": "slank"
      }
    ]
  }
}
`

const TestAgentNoDatabaseOutput2 = `{
  "users": {
    "user": [
      {
        "name": "hiroki"
      }
    ]
  }
}
`

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

const TestAgentLoadDatabaseOutput1 = `{
  "users": {
    "user": [
      {
        "age": 22,
        "name": "hiroki"
      },
      {
        "age": 30,
        "name": "slank"
      }
    ]
  }
}
`

const TestAgentLoadDatabaseOutput2 = `{
  "users": {
    "user": [
      {
        "age": 22,
        "name": "hiroki"
      },
      {
        "age": 30,
        "name": "slank"
      },
      {
        "age": 28,
        "name": "shirokura",
        "projects": [
          {
            "name": "mfplane"
          }
        ]
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
