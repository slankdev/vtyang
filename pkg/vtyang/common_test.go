package vtyang

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/slankdev/vtyang/pkg/util"
)

const (
	agentTestDefaultLogFile = "/tmp/run/vtyang/vtyang.log"
)

type TestCase struct {
	YangPath       string
	RuntimePath    string
	LogFile        string
	Inputs         []string
	OutputString   string
	OutputFile     string
	InitConfigFile string
}

func executeTestCase(t *testing.T, tc *TestCase) {
	// Preparation
	GlobalOptRunFilePath = tc.RuntimePath
	if util.FileExists(getDatabasePath()) {
		if err := os.Remove(getDatabasePath()); err != nil {
			t.Fatal(err)
		}
	}

	if tc.InitConfigFile != "" {
		in, err := os.ReadFile(tc.InitConfigFile)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(getDatabasePath(), in, 0644); err != nil {
			t.Error(err)
		}
	}

	// NOTE(slankdev): set as default
	if tc.LogFile == "" {
		tc.LogFile = agentTestDefaultLogFile
	}

	// Initializing Agent
	if err := InitAgent(AgentOpts{
		RuntimePath: tc.RuntimePath,
		YangPath:    []string{tc.YangPath},
		LogFile:     tc.LogFile,
	}); err != nil {
		t.Fatal(err)
	}

	// Execute Test commands
	buf := setStdoutWithBuffer()
	for idx, input := range tc.Inputs {
		t.Logf("Testcase[%d] executing %s", idx, input)
		getCommandNodeCurrent().executeCommand(input)
	}
	result := buf.String()

	// Check output
	out, err := os.ReadFile(tc.OutputFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != result {
		t.Errorf("Unexpected output")
		p := fmt.Sprintf("/tmp/test_fail_output_%s", t.Name())
		if err := os.WriteFile(p+"_expected.txt", out, os.ModePerm); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p+"_result.txt", []byte(result), os.ModePerm); err != nil {
			t.Fatal(err)
		}
		t.Errorf("expect(len=%d) %s_expected.txt", len(string(out)), p)
		t.Errorf("result(len=%d) %s_result.txt", len(result), p)
		t.Errorf("KINDLY_CLI diff -u %s_expected.txt %s_result.txt", p, p)
		t.Errorf("KINDLY_CLI vim -O %s_result.txt %s",
			p, path.Join("./pkg/vtyang", tc.OutputFile))
		t.Fatal("quiting test with FAILED result")
	}

	// Return
	t.Logf("Testcase output check is succeeded")
}
