package vtyang

import (
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
	testcases := []TestCaseForTestAgent{
		{
			Inputs: []string{
				"configure",
				"set isis instance 1 default description area1-default-hoge",
				"set isis instance 1 vrf0 description area1-vrf0-hoge",
				"set isis instance 2 vrf0 description area2-vrf0-hoge",
				"set isis instance 1 vrf0 description area1-vrf0-fuga",
				"set isis instance 1 default area-address 10.0000.0000.0000.0000.0000.0000.0000.0000.0000.00",
				"commit",
				"do show running-config-frr",
			},
			Output: `{

			}`,
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
		eq, err := util.DeepEqualJSON(result, tc.Output)
		if err != nil {
			t.Fatal(err)
		}
		if !eq {
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
