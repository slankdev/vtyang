package vtyang

import (
	"fmt"
	"sort"
	"strings"

	"github.com/openconfig/goyang/pkg/yang"
)

func getCommandOperState(modules *yang.Modules) *CompletionNode {
	fullnames := []string{}
	for fullname := range modules.Modules {
		if !strings.Contains(fullname, "@") {
			fullnames = append(fullnames, fullname)
		}
	}
	sort.Strings(fullnames)
	child := []*CompletionNode{}
	for _, fullname := range fullnames {
		m := modules.Modules[fullname]
		entnames := []string{}
		for entname := range yang.ToEntry(m).Dir {
			entnames = append(entnames, entname)
		}
		sort.Strings(entnames)
		for _, entname := range entnames {
			e := yang.ToEntry(m).Dir[entname]
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
				switch {
				case ee.IsChoice():
					for _, ee2 := range ee.Dir {
						for _, ee3 := range ee2.Dir {
							if !ee3.ReadOnly() && ee3.RPC == nil {
								wildcardNode.Childs = append(wildcardNode.Childs,
									resolveCompletionNodeOperState(ee3, depth+1))
								sort.Slice(wildcardNode.Childs,
									func(i, j int) bool {
										return wildcardNode.Childs[i].Name < wildcardNode.Childs[j].Name
									})
							}
						}
					}
				default:
					if nn := resolveCompletionNodeOperState(ee, depth+1); nn != nil {
						wildcardNode.Childs = append(wildcardNode.Childs, nn)
					}
				}
			}
		}
		n.Childs = append(n.Childs, wildcardNode)
		sort.Slice(n.Childs, func(i, j int) bool { return n.Childs[i].Name < n.Childs[j].Name })

	case e.IsLeaf():
		child := &CompletionNode{
			Name:   "VALUE",
			Childs: []*CompletionNode{newCR()},
		}
		n.Childs = append(n.Childs, child)
		sort.Slice(n.Childs, func(i, j int) bool { return n.Childs[i].Name < n.Childs[j].Name })

	default:
		childs := []*CompletionNode{}
		for _, ee := range e.Dir {
			switch {
			case ee.IsChoice():
				for _, ee2 := range ee.Dir {
					for _, ee3 := range ee2.Dir {
						if !ee3.ReadOnly() && ee3.RPC == nil {
							childs = append(childs, resolveCompletionNodeOperState(ee3, depth+1))
						}
					}
				}
			default:
				childs = append(childs, resolveCompletionNodeOperState(ee, depth+1))
			}
		}
		n.Childs = childs
		sort.Slice(n.Childs, func(i, j int) bool { return n.Childs[i].Name < n.Childs[j].Name })
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
