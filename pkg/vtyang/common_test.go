package vtyang

import (
	"fmt"
	"os"
	"testing"

	"github.com/slankdev/vtyang/pkg/util"
)

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
