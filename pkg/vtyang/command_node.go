package vtyang

import (
	"fmt"
	"strings"

	"github.com/k0kubun/pp"
	"github.com/openconfig/goyang/pkg/yang"
)

type CliMode int

const (
	CliModeView CliMode = iota
	CliModeConfigure
)

type Command struct {
	m string
	f func(args []string)
}

type CommandNode struct {
	mode     CliMode
	commands []Command
	tree     CompletionTree
}

var cliMode CliMode = CliModeView
var commandnodes map[CliMode]*CommandNode

func GetCommandNode(mode CliMode) *CommandNode {
	if commandnodes == nil {
		commandnodes = map[CliMode]*CommandNode{}
	}

	cn, ok := commandnodes[mode]
	if !ok {
		ncn := CommandNode{}
		ncn.mode = mode
		ncn.tree = CompletionTree{}
		ncn.tree.Root = &CompletionNode{}
		commandnodes[mode] = &ncn
		cn = commandnodes[mode]
	}
	return cn
}

func GetCommandNodeCurrent() *CommandNode {
	return GetCommandNode(cliMode)
}

func (cn *CommandNode) ExecuteCommand(cli string) {
	args := strings.Fields(cli)
	notfound := true
	for _, cmd := range cn.commands {
		if matchArgs(args, cmd.m) {
			cmd.f(args)
			notfound = false
			break
		}
	}
	if notfound {
		fmt.Printf("Error: command %s not found\n", args[0])
	}
}

func InstallCommandsDefault(mode CliMode) {
	InstallCommand(mode, "list", []string{"List cli nodes"}, func(arg []string) {
		cn := GetCommandNodeCurrent()
		for _, cmd := range cn.commands {
			fmt.Printf("%s\n", cmd.m)
		}
	})

	InstallCommand(mode, "quit", []string{"Quit system"}, func(arg []string) {
		switch cliMode {
		case CliModeView:
			exit = true
		case CliModeConfigure:
			cliMode = CliModeView
			dbm.candidateRoot = nil
		}
	})

	InstallCommand(mode, "show cli-tree", []string{
		"Display information",
		"Display completion tree",
	}, func(arg []string) {
		tree := GetCommandNodeCurrent().tree
		pp.Println(tree)
	})
}

func InstallCommand(mode CliMode, match string, helps []string,
	f func(args []string)) {
	cn := GetCommandNode(mode)
	cn.commands = append(cn.commands, Command{m: match, f: f})

	args := strings.Fields(match)
	if len(args) != len(helps) {
		panic(fmt.Sprintf("ERROR %s len(helps)=%d", match, len(helps)))
	}

	node := cn.tree.Root
	var nn *CompletionNode
	for ; len(args) != 0; args, helps = args[1:], helps[1:] {
		for idx := range node.Childs {
			child := node.Childs[idx]
			if child.Name == args[0] {
				node = child
				goto end
			}
		}

		nn = new(CompletionNode)
		nn.Name = args[0]
		nn.Description = helps[0]
		if len(args) == 1 {
			nn.Childs = []*CompletionNode{{Name: "<cr>"}}
		}
		node.Childs = append(node.Childs, nn)
		node = node.Childs[len(node.Childs)-1]
	end:
	}
}

func InstallCommandYang(mode CliMode, match string, helps []string,
	callbackFunc func(args []string),
	completionFunc func(ent *yang.Entry) *CompletionNode) {
	InstallCommand(mode, match, helps, callbackFunc)

	root := DigNodeOrDie(mode, strings.Fields(match))
	for _, e := range dbm.YangEntries() {
		n := completionFunc(e)
		if n != nil {
			root.Childs = append(root.Childs, n)
		}
	}
}
