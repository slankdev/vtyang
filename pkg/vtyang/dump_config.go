package vtyang

import (
	"fmt"

	"github.com/openconfig/goyang/pkg/yang"
)

func getCommandConfig(modules *yang.Modules) *CompletionNode {
	child := []*CompletionNode{}
	for _, m := range modules.Modules {
		for _, e := range yang.ToEntry(m).Dir {
			if !e.ReadOnly() && e.RPC == nil {
				child = append(child, resolveCompletionNodeConfig(e, 0))
			}
		}
	}
	return &CompletionNode{
		Childs: []*CompletionNode{
			{
				Name:   "show",
				Childs: child,
			},
			{
				Name:   "set",
				Childs: child,
			},
			{
				Name:   "delete",
				Childs: child,
			},
		},
	}
}

func resolveCompletionNodeConfig(e *yang.Entry, depth int) *CompletionNode {
	n := CompletionNode{}
	n.Name = e.Name
	n.Description = e.Description

	switch {
	case e.IsList():
		wildcardNode := &CompletionNode{
			Name:   "NAME",
			Childs: []*CompletionNode{newCR()},
		}
		for _, ee := range e.Dir {
			if ee.Name != e.Key {
				nn := resolveCompletionNodeConfig(ee, depth+1)
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
		for _, ee := range e.Dir {
			if !ee.ReadOnly() && ee.RPC == nil {
				childs = append(childs, resolveCompletionNodeConfig(ee, depth+1))
			}
		}
		n.Childs = childs
	}
	return &n
}

func getCommandCallbackConfig(_ *yang.Modules) []Command {
	return []Command{
		{
			m: "show",
			f: func(args []string) {
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
				fmt.Fprintln(stdout, stdout, stdout, node.String())
			},
		},
		{
			m: "set",
			f: func(args []string) {
				xpath, valueStr, err := ParseXPathArgs(dbm, args[1:], true)
				if err != nil {
					fmt.Fprintf(stdout, "Error: %s\n", err.Error())
					return
				}
				if _, err := dbm.SetNode(xpath, valueStr); err != nil {
					fmt.Fprintf(stdout, "Error: %s\n", err.Error())
					return
				}
			},
		},
		{
			m: "delete",
			f: func(args []string) {
				xpath, _, err := ParseXPathArgs(dbm, args[1:], true)
				if err != nil {
					fmt.Fprintf(stdout, "Error: %s\n", err.Error())
					return
				}
				if err := dbm.DeleteNode(xpath); err != nil {
					fmt.Fprintf(stdout, "Error: %s\n", err.Error())
					return
				}
			},
		},
	}
}

func getViewCommandConfig(modules *yang.Modules) *CompletionNode {
	child := []*CompletionNode{}
	for _, m := range modules.Modules {
		for _, e := range yang.ToEntry(m).Dir {
			if !e.ReadOnly() && e.RPC == nil {
				child = append(child, resolveCompletionNodeConfig(e, 0))
			}
		}
	}
	return &CompletionNode{
		Childs: []*CompletionNode{
			{
				Name: "show",
				Childs: []*CompletionNode{
					{
						Name:   "running-config",
						Childs: child,
					},
				},
			},
		},
	}
}

func getViewCommandCallbackConfig(_ *yang.Modules) []Command {
	return []Command{
		{
			m: "show running-config",
			f: func(args []string) {
				xpath, _, err := ParseXPathArgs(dbm, args[2:], false)
				if err != nil {
					fmt.Fprintf(stdout, "Error: %s\n", err.Error())
					return
				}
				node, err := dbm.GetNode(xpath)
				if err != nil {
					fmt.Fprintf(stdout, "Error: %s\n", err.Error())
					return
				}
				fmt.Fprintln(stdout, node.String())
			},
		},
	}
}
