package vtyang

import (
	"fmt"

	"github.com/openconfig/goyang/pkg/yang"
)

func getCommandOperState(modules *yang.Modules) *CompletionNode {
	child := []*CompletionNode{}
	for _, m := range modules.Modules {
		for _, e := range yang.ToEntry(m).Dir {
			child = append(child, resolveCompletionNodeOperState(e, 0))
		}
	}
	return &CompletionNode{
		Childs: []*CompletionNode{
			{
				Name:   "show",
				Childs: child,
			},
		},
	}
}

func resolveCompletionNodeOperState(e *yang.Entry, depth int) *CompletionNode {
	n := CompletionNode{}
	n.Name = e.Name
	n.Description = e.Description

	switch {
	case e.IsList():
		wildcardNode := &CompletionNode{}
		wildcardNode.Name = "NAME"
		wildcardNode.Childs = []*CompletionNode{newCR()}
		for _, ee := range e.Dir {
			if ee.Name != e.Key {
				if nn := resolveCompletionNodeConfig(ee, depth+1); nn != nil {
					wildcardNode.Childs = append(wildcardNode.Childs, nn)
				}
			}
		}
		n.Childs = append(n.Childs, wildcardNode)

	case e.IsLeaf():
		child := &CompletionNode{
			Name:   "VALUE",
			Childs: []*CompletionNode{newCR()},
		}
		n.Childs = append(n.Childs, child)

	default:
		childs := []*CompletionNode{}
		for _, ee := range e.Dir {
			childs = append(childs, resolveCompletionNodeOperState(ee, depth+1))
		}
		n.Childs = childs
	}
	return &n
}

func getCommandCallbackOperState(_ *yang.Modules) []Command {
	return []Command{
		{
			m: "show",
			f: func(args []string) {
				if len(args) < 2 {
					fmt.Fprintf(stdout, "usage:\n")
					return
				}
				xpath, _, err := ParseXPathArgs(dbm, args[1:], false)
				if err != nil {
					fmt.Fprintf(stdout, "Error: %s\n", err.Error())
					return
				}
				node, err := dbm.GetNode(xpath)
				if err != nil {
					fmt.Fprintf(stdout, "Error: %s\n", err.Error())
					return
				}
				if node == nil {
					fmt.Fprintf(stdout, "Not Found\n")
					return
				}
				fmt.Fprintln(stdout, node.String())
			},
		},
	}
}
