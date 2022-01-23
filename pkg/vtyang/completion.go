package vtyang

import (
	"fmt"
	"log"
	"strings"
)

type CompletionTree struct {
	Root *CompletionNode
}

type CompletionNode struct {
	Name        string
	Description string
	Childs      []*CompletionNode
}

func (n *CompletionNode) String() string {
	name := "<root>"
	if n.Name != "" {
		name = n.Name
	}
	return fmt.Sprintf("%s", name)
}

func (t *CompletionTree) Completion(line string, pos int) []*CompletionNode {
	line = line[:pos]
	args := strings.Fields(line)
	if len(args) == 0 {
		args = append(args, "")
	} else if line[len(line)-1] == ' ' {
		args = append(args, "")
	}
	log.Printf("DEBUG: line=\"%s\" pos=%d, args=%d:%+v", line, pos, len(args),
		args)

	search := func(nodes []*CompletionNode, str string) []*CompletionNode {
		result := []*CompletionNode{}
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

	tree := GetCommandNodeCurrent().tree
	var pivot *CompletionNode = tree.Root
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
			pivot = candidates[0]
		case n >= 1:
			pivot = candidates[0]
		}
	}

	return nil
}

func DigNode(node *CompletionNode, query []string) *CompletionNode {
	if len(query) == 0 {
		return node
	}

	for idx, _ := range node.Childs {
		child := node.Childs[idx]
		if child.Name == query[0] {
			return DigNode(child, query[1:])
		}
	}
	return nil

}

func DigNodeOrDie(mode CliMode, query []string) *CompletionNode {
	root := DigNode(GetCommandNode(mode).tree.Root, query)
	if root == nil {
		panic(fmt.Sprintf("Notfound %+v", query))
	}
	return root
}

func completer(line string, pos int) (string, []string, string) {
	tree := GetCommandNodeCurrent().tree
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

func completionLister(line string, pos int) {
	tree := GetCommandNodeCurrent().tree
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
