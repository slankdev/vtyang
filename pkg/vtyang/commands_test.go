package vtyang

import (
	"reflect"
	"testing"

	"github.com/k0kubun/pp"
	"github.com/pkg/errors"
)

func TestInstallCompletionTree(t *testing.T) {
	root := &CompletionNode{
		Name: "",
		Childs: []*CompletionNode{
			{
				Name: "show",
				Childs: []*CompletionNode{
					{
						Name: "configuration",
						Childs: []*CompletionNode{
							newCR(),
						},
					},
				},
			},
		},
	}
	inputRoot := &CompletionNode{
		Childs: []*CompletionNode{
			{
				Name: "show",
				Childs: []*CompletionNode{
					{
						Name: "running-config",
						Childs: []*CompletionNode{
							newCR(),
						},
					},
				},
			},
		},
	}
	expect := &CompletionNode{
		Name: "",
		Childs: []*CompletionNode{
			{
				Name: "show",
				Childs: []*CompletionNode{
					{
						Name: "configuration",
						Childs: []*CompletionNode{
							newCR(),
						},
					},
					{
						Name: "running-config",
						Childs: []*CompletionNode{
							newCR(),
						},
					},
				},
			},
		},
	}
	if err := installCompletionTree(root, inputRoot); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(root, expect) {
		pp.Println("expect", expect)
		pp.Println("result", root)
		t.Errorf("missmatch")
	}
}

type TestDoCompletionTestCase struct {
	in  string
	out CompletionResult
}

func executeDoCompletionTestCase(testcase []TestDoCompletionTestCase, idx int) error {
	tc := testcase[idx]
	result := doCompletion(tc.in, len(tc.in))
	result.ResolvedXPath = nil
	if !reflect.DeepEqual(result, tc.out) {
		pp.Println("expect", tc.out)
		pp.Println("result", result)
		return errors.Errorf("diff tc[%d] \"%s\"", idx, tc.in)
	}
	return nil
}

func TestDoCompletion01(t *testing.T) {
	// Init agent
	if err := InitAgent(AgentOpts{
		LogFile:     agentTestDefaultLogFile,
		RuntimePath: "/tmp/run/vtyang",
		YangPath:    []string{"./testdata/yang/basic"},
	}); err != nil {
		t.Fatal(err)
	}
	getCommandNodeCurrent().executeCommand("configure")

	// Testcase Decration
	testcases := []TestDoCompletionTestCase{
		{
			in: "set values",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "values"},
				},
			},
		},
		{
			in: "set values ",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "afi"},
					{Word: "bool"},
					{Word: "crypto"},
					{Word: "decimal"},
					{Word: "i08"},
					{Word: "i16"},
					{Word: "i32"},
					{Word: "i64"},
					{Word: "ipv4-address"},
					{Word: "ipv6-address"},
					{Word: "items"},
					{Word: "month"},
					{Word: "month-str"},
					{Word: "month-union"},
					{Word: "name"},
					{Word: "percentage"},
					{Word: "transport-proto"},
					{Word: "u08"},
					{Word: "u16"},
					{Word: "u32"},
					{Word: "u64"},
					{Word: "union-list"},
					{Word: "union-multiple"},
				},
			},
		},
		{
			in: "set values cry",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "crypto"},
				},
			},
		},
		{
			// TODO(slankdev): Okashii This behavior
			// out.InvalidArg must be true
			in: "set values cry ",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "VALUE"},
				},
			},
		},
		{
			in: "set values crypto",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "crypto"},
				},
			},
		},
		{
			in: "set values crypto ",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "aes"},
					{Word: "des3"},
				},
			},
		},
		{
			in: "set values crypto ae",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "aes"},
				},
			},
		},
		{
			in: "set values crypto aes",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "aes"},
				},
			},
		},
		{
			in: "set values crypto aes ",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "<cr>"},
				},
			},
		},
	}

	// Executes
	for idx := range testcases {
		t.Logf("execute tc[%d] \"%s\"", idx, testcases[idx].in)
		if err := executeDoCompletionTestCase(testcases, idx); err != nil {
			t.Errorf("%s\n", err)
		}
	}
}
