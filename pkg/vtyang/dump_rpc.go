package vtyang

import (
	"fmt"
	"strings"

	"github.com/openconfig/goyang/pkg/yang"
)

func getCommandRPC(modules *yang.Modules) *CompletionNode {
	child := []*CompletionNode{}
	for fullname, m := range modules.Modules {
		if strings.Contains(fullname, "@") {
			continue
		}
		for _, e := range yang.ToEntry(m).Dir {
			if e.RPC != nil {
				child = append(child, resolveCompletionNodeRPC(e, 0))
			}
		}
	}
	return &CompletionNode{
		Childs: []*CompletionNode{
			{
				Name:   "rpc",
				Childs: child,
			},
		},
	}
}

func resolveCompletionNodeRPC(e *yang.Entry, depth int) *CompletionNode {
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
				nn := resolveCompletionNodeRPC(ee, depth+1)
				if nn != nil {
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
		if e.RPC != nil {
			if e.RPC.Input != nil {
				for _, ee := range e.RPC.Input.Dir {
					childs = append(childs, resolveCompletionNodeRPC(ee, depth+1))
				}
			}
		}
		for _, ee := range e.Dir {
			childs = append(childs, resolveCompletionNodeRPC(ee, depth+1))
		}
		if len(childs) == 0 {
			childs = append(childs, newCR())
		}
		n.Childs = childs
	}
	return &n
}

func getCommandCallbackRPC(_ *yang.Modules) []Command {
	return []Command{
		{
			m: "rpc",
			f: func(args []string) {
				fmt.Fprintln(stdout, "NOT IMPLEMENTED")
			},
		},
	}
}
