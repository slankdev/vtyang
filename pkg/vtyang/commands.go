package vtyang

import (
	"fmt"

	"github.com/k0kubun/pp"
	"github.com/openconfig/goyang/pkg/yang"
)

func C(dbm *DatabaseManager, args []string) {
	if len(args) < 3 {
		fmt.Printf("usage:\n")
		return
	}

	mod, xpath := dbm.CraftXPath(args[2:])
	node, err := dbm.GetNode(mod, xpath)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return
	}

	pp.Println(node)
}

func (dbm *DatabaseManager) CraftXPath(args []string) (string, string) {
	ents := dbm.DumpEntries()
	for _, e := range ents {
		if e.Name == args[0] {
			xpath := dbm.CraftXPathRc(args[1:], e, "")
			return e.Name, xpath
		}
	}
	return "???", "???"
}

func (dbm *DatabaseManager) CraftXPathRc(args []string, e *yang.Entry, xpath string) string {
	if len(args) == 0 {
		return xpath
	}

	word := args[0]
	//fmt.Printf("DEBUG word=%s xpath=%s (%+v)\n", word, xpath, e.Kind)

	for _, child := range e.Dir {
		if child.Name == word {
			switch {
			case child.IsLeaf():
				xpath = xpath + "/" + word
				return xpath
			case child.IsList():
				if len(args) < 2 {
					fmt.Printf("invalid xpath %s/%s\n", xpath, word)
					return "nil"
				}
				xpath = fmt.Sprintf("%s/%s['%s'='%s']", xpath, word, child.Key, args[1])
				// fmt.Printf("DEBUG X=%s\n", xpath)
				return dbm.CraftXPathRc(args[2:], child, xpath)
			default:
				xpath = xpath + "/" + word
				return dbm.CraftXPathRc(args[1:], child, xpath)
			}
		}
	}

	return "<??>"
}
