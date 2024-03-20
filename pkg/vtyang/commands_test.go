package vtyang

import (
	"reflect"
	"testing"

	"github.com/k0kubun/pp"
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
