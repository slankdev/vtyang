package vtyang

import (
	"encoding/json"
	"fmt"
	"sort"
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
			c := resolveCompletionNodeConfig(e, 0, m.Name)
			found := false
			for idx := range child {
				if child[idx].Name == c.Name {
					child[idx] = merge(child[idx], c)
					found = true
					break
				}
			}
			if !found {
				child = append(child, resolveCompletionNodeConfig(e, 0, m.Name))
			}
		}
	}
	sort.Slice(child, func(i, j int) bool { return child[i].Name < child[j].Name })
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
				sort.Slice(top.Childs, func(i, j int) bool { return top.Childs[i].Name < top.Childs[j].Name })
			}
		}
		for _, ee := range e.Dir {
			if ee.Name != e.Key {
				switch {
				case ee.IsChoice():
					for _, ee2 := range ee.Dir {
						for _, ee3 := range ee2.Dir {
							if !ee3.ReadOnly() && ee3.RPC == nil {
								tail.Childs = append(tail.Childs,
									resolveCompletionNodeConfig(ee3, depth+1, modName))
								sort.Slice(tail.Childs,
									func(i, j int) bool {
										return tail.Childs[i].Name < tail.Childs[j].Name
									})
							}
						}
					}
				default:
					nn := resolveCompletionNodeConfig(ee, depth+1, modName)
					if nn != nil {
						tail.Childs = append(tail.Childs, nn)
						sort.Slice(tail.Childs, func(i, j int) bool { return tail.Childs[i].Name < tail.Childs[j].Name })
					}
				}
			}
		}
		n.Childs = append(n.Childs, top)
		sort.Slice(n.Childs, func(i, j int) bool { return n.Childs[i].Name < n.Childs[j].Name })

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
				switch {
				case ee.IsChoice():
					for _, ee2 := range ee.Dir {
						for _, ee3 := range ee2.Dir {
							if !ee3.ReadOnly() && ee3.RPC == nil {
								childs = append(childs, resolveCompletionNodeConfig(ee3, depth+1, modName))
							}
						}
					}
				default:
					childs = append(childs, resolveCompletionNodeConfig(ee, depth+1, modName))

				}
			}
		}
		sort.Slice(childs, func(i, j int) bool { return childs[i].Name < childs[j].Name })
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

func merge(n1, n2 *CompletionNode) *CompletionNode {
	if n1.Name != n2.Name {
		panic("ASSERTION")
	}
	new := CompletionNode{
		Name:        n1.Name,
		Description: n1.Description,
	}
	new.Modules = append(new.Modules, n1.Modules...)
	new.Modules = append(new.Modules, n2.Modules...)
	sort.Slice(new.Modules, func(i, j int) bool { return new.Modules[i] < new.Modules[j] })

	// Only exist in n1
	for idx1 := range n1.Childs {
		found := false
		for idx2 := range n2.Childs {
			if n1.Childs[idx1].Name == n2.Childs[idx2].Name {
				found = true
				break
			}
		}
		if !found {
			new.Childs = append(new.Childs, n1.Childs[idx1])
		}
	}

	// Only exist in n2
	for idx2 := range n2.Childs {
		found := false
		for idx1 := range n1.Childs {
			if n1.Childs[idx1].Name == n2.Childs[idx2].Name {
				found = true
				break
			}
		}
		if !found {
			new.Childs = append(new.Childs, n2.Childs[idx2])
		}
	}

	// Duplicate in n1 and n2
	for idx1 := range n1.Childs {
		for idx2 := range n2.Childs {
			if n1.Childs[idx1].Name == n2.Childs[idx2].Name {
				n := merge(n1.Childs[idx1], n2.Childs[idx2])
				new.Childs = append(new.Childs, n)
				break
			}
		}
	}

	sort.Slice(new.Childs, func(i, j int) bool { return new.Childs[i].Name < new.Childs[j].Name })
	return &new
}

func getViewCommandConfig(modules *yang.Modules) *CompletionNode {
	child := []*CompletionNode{}
	for fullname, m := range modules.Modules {
		if strings.Contains(fullname, "@") {
			continue
		}
		for _, e := range yang.ToEntry(m).Dir {
			if !e.ReadOnly() && e.RPC == nil {
				c := resolveCompletionNodeConfig(e, 0, m.Name)
				found := false
				for idx := range child {
					if child[idx].Name == c.Name {
						child[idx] = merge(child[idx], c)
						found = true
						break
					}
				}
				if !found {
					child = append(child, resolveCompletionNodeConfig(e, 0, m.Name))
				}
			}
		}
	}
	sort.Slice(child, func(i, j int) bool { return child[i].Name < child[j].Name })
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
					{
						Name:   "running-config-raw",
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
				node := &dbm.root
				filteredNode, err := filterDbWithModule(node, "frr-isisd")
				if err != nil {
					fmt.Fprintf(stdout, "Error: %s\n", err.Error())
					return
				}
				fmt.Fprintln(stdout, filteredNode.String())
			},
		},
		{
			m: "show running-config-raw",
			f: func(args []string) {
				node := &dbm.root
				out, err := json.MarshalIndent(node, "", "  ")
				if err != nil {
					fmt.Fprintf(stdout, "Error: %s\n", err.Error())
					return
				}
				fmt.Fprintln(stdout, string(out))
			},
		},
	}
}
