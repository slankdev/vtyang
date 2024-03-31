package vtyang

import (
	"fmt"
	"strings"

	"github.com/openconfig/goyang/pkg/yang"
)

func getCommandConfig(modules *yang.Modules) *CompletionNode {
	child := []*CompletionNode{}
	for fullname, m := range modules.Modules {
		if strings.Contains(fullname, "@") {
			continue
		}
		for _, e := range yang.ToEntry(m).Dir {
			if e.ReadOnly() {
				continue
			}
			if e.RPC != nil {
				continue
			}
			if e.Kind == yang.NotificationEntry {
				continue
			}
			child = append(child, resolveCompletionNodeConfig(e, 0, m.Name))
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

func resolveCompletionNodeConfig(e *yang.Entry, depth int, modName string) *CompletionNode {
	n := CompletionNode{}
	n.Name = e.Name
	n.Modules = []string{modName}

	// TODO(slankdev):
	//
	// According to IETF yang, description seems to be too long for a CLI help
	// string. We will discuss whether we should design some kind of extension
	// for appropriate cli help strings like tailf:foo.
	//
	// n.Description = e.Description

	switch {
	case e.IsList():
		var top *CompletionNode = nil
		var tail *CompletionNode = nil
		for _, word := range strings.Fields(e.Key) {
			tail = &CompletionNode{
				Name:        "NAME",
				Description: word,
				Childs:      []*CompletionNode{newCR()},
			}
			if top == nil {
				top = tail
			} else {
				top.Childs = append(top.Childs, tail)
			}
		}
		for _, ee := range e.Dir {
			if ee.Name != e.Key {
				nn := resolveCompletionNodeConfig(ee, depth+1, modName)
				if nn != nil {
					tail.Childs = append(tail.Childs, nn)
				}
			}
		}
		n.Childs = append(n.Childs, top)

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
				childs = append(childs, resolveCompletionNodeConfig(ee, depth+1, modName))
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
	for fullname, m := range modules.Modules {
		if strings.Contains(fullname, "@") {
			continue
		}
		for _, e := range yang.ToEntry(m).Dir {
			if !e.ReadOnly() && e.RPC == nil {
				child = append(child, resolveCompletionNodeConfig(e, 0, m.Name))
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
					{
						Name:   "running-config-frr",
						Childs: []*CompletionNode{newCR()},
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
		{
			m: "show running-config-frr",
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
				filteredNode, err := filterDbWithModule(node, "frr-isisd")
				if err != nil {
					fmt.Fprintf(stdout, "Error: %s\n", err.Error())
					return
				}
				fmt.Fprintln(stdout, filteredNode.String())
			},
		},
	}
}
