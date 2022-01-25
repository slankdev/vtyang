package vtyang

import (
	"fmt"
	"log"
	"strings"

	"github.com/k0kubun/pp"
	"github.com/slankdev/vtyang/pkg/liner"
	"github.com/spf13/cobra"
)

const (
	// Question mark '?'
	QUESTION_MARK rune = 63
)

type CompletionTree struct {
	Root CompletionNode
}

type CompletionNode struct {
	Name        string
	Description string
	Childs      []CompletionNode
	Level       int
}

func (n CompletionNode) String() string {
	name := "<root>"
	if n.Name != "" {
		name = n.Name
	}
	return fmt.Sprintf("%d:%s", n.Level, name)
}

func (t CompletionTree) Completion(line string, pos int) []CompletionNode {
	line = line[:pos]
	args := strings.Fields(line)
	if len(args) == 0 {
		args = append(args, "")
	} else if line[len(line)-1] == ' ' {
		args = append(args, "")
	}
	log.Printf("DEBUG: line=\"%s\" pos=%d, args=%d:%+v", line, pos, len(args),
		args)

	search := func(nodes []CompletionNode, str string) []CompletionNode {
		result := []CompletionNode{}
		for _, node := range nodes {
			switch node.Name {
			case "NAME":
				result = append(result, node)

			// // TODO(slankdev)
			// case "INTEGER":
			// 	if _, err := strconv.Atoi(str); err == nil {
			// 		result = append(result, node)
			// 	} else if str == "" {
			// 		result = append(result, node)
			// 	}

			default:
				log.Printf("HasPrefix(%s,%s)\n", node.Name, str)
				if strings.HasPrefix(node.Name, str) {
					result = append(result, node)
				}
			}
		}
		return result
	}

	var pivot *CompletionNode = &tree.Root
	last := len(args) - 1
	for i, arg := range args {
		if i == last {
			log.Printf("loop (%d) last\n", i)
		} else {
			log.Printf("loop (%d)\n", i)
		}
		log.Printf("loop (%d) %s", i, pivot.String())

		candidates := search(pivot.Childs, arg)
		n := len(candidates)
		log.Printf("loop (%d) match %d candidates\n", i, n)

		if i == last {
			return candidates
		}

		switch {
		case n == 0:
			return nil
		case n == 1:
			pivot = &candidates[0]
		case n >= 1:
			pivot = &candidates[0]
		}
	}

	return nil
}

// func init() {
// 	InstallCommand("show operational-data",
//                           "Display information\n"+
// 			  "Display yang modules\n",
// 			  func (args []string) error {
// 				dbm.Dump()
// 			  },
// 			func(...) {
// 				...
// 			})
// }

func DigNode(node *CompletionNode, query []string) *CompletionNode {
	if len(query) == 0 {
		return node
	}

	for idx, _ := range node.Childs {
		child := &node.Childs[idx]
		if child.Name == query[0] {
			return DigNode(child, query[1:])
		}
	}
	return nil

}

func completer(line string, pos int) (string, []string, string) {
	nodes := tree.Completion(line, pos)
	if len(nodes) == 0 {
		return line[:pos], nil, line[pos:]
	}

	names := []string{}
	for _, node := range nodes {
		names = append(names, node.Name)
	}

	pre := line[:pos]
	if len(pre) > 0 && pre[len(pre)-1] != ' ' {
		words := strings.Fields(pre)
		last := len(words) - 1
		words = append(words[:last], words[last+1:]...)
		pre = ""
		for _, word := range words {
			pre += word
			pre += " "
		}
	}

	if len(names) == 1 && names[0] == "<cr>" {
		return pre, nil, line[pos:]
	}

	if len(names) > 0 && names[len(names)-1] == "NAME" {
		return pre, nil, line[pos:]
	}

	if len(names) == 1 {
		names[0] += " "
		return pre, names, line[pos:]
	}

	return pre, names, line[pos:]
}

func binder(line string, pos int) {
	nodes := tree.Completion(line, pos)
	if len(nodes) == 0 {
		fmt.Printf("\n%% Invalid input detected\n")
		return
	}

	longestnamelen := 0
	for _, node := range nodes {
		if len(node.Name) > longestnamelen {
			longestnamelen = len(node.Name)
		}
	}
	padding := func(str string, maxlen int) string {
		retStr := ""
		for i := 0; i < maxlen-len(str); i++ {
			retStr += " "
		}
		return retStr
	}

	fmt.Printf("\nPossible Completions:\n")
	for _, node := range nodes {
		fmt.Printf("  %s%s  %s\n", node.Name, padding(node.Name, longestnamelen),
			node.Description)
	}
}

func match(args []string, matchStr string) bool {
	matchArgs := strings.Fields(matchStr)
	if len(matchArgs) > len(args) {
		log.Printf("Unmatch %s v.s. %s\n", args, matchStr)
		return false
	}

	for i, _ := range matchArgs {
		if matchArgs[i] != args[i] {
			log.Printf("Unmatch %s v.s. %s\n", args, matchStr)
			return false
		}
	}

	log.Printf("Match %s v.s. %s\n", args, matchStr)
	return true
}

func agentMain(cmd *cobra.Command, args []string) error {
	dbm = NewDatabaseManager()
	dbm.LoadYangModuleOrDie("./yang")
	setCompletionTreeForCommandShowOperationalData()
	setCompletionTreeForCommandSet()

	ErrorOnDie(dbm.Create("account", "/users/user['name'='hiroki']/projects['name'='tennis']/finished"))

	line := liner.NewLiner()
	defer line.Close()
	line.SetCtrlCAborts(true)
	line.SetWordCompleter(completer)
	line.SetTabCompletionStyle(liner.TabPrints)
	line.SetBinder(QUESTION_MARK, binder)

	for {
		if name, err := line.Prompt("vtyang# "); err == nil {
			line.AppendHistory(name)
			name = strings.TrimSpace(name)
			args := strings.Fields(name)
			if len(args) == 0 {
				continue
			}

			switch {
			case match(args, "set"):
				fmt.Printf("not implemented\n")
			case match(args, "show running-config"):
				pp.Println(dbm.db.root)
			case match(args, "show yang-modules"):
				dbm.Dump()
			case match(args, "show cli-tree"):
				pp.Println(tree)
			case match(args, "show operational-data"):
				pp.Println(DummyDBRoot)
				C(dbm, args)
			case match(args, "show"):
				fmt.Printf("not implemented\n")
			default:
				fmt.Printf("Error: command %s not found\n", args[0])
			}

		} else if err == liner.ErrPromptAborted {
			log.Print("aborted")
			break
		} else {
			log.Print("error reading line: ", err)
		}
	}

	return nil
}
