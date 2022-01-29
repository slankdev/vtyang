package vtyang

import (
	"fmt"
	"strings"
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
		commandnodes[mode] = &ncn
		cn = commandnodes[mode]
		InstallCommand(mode, "list", listCallback)
		InstallCommand(mode, "quit", quitCallback)
	}
	return cn
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

func InstallCommand(mode CliMode, match string, f func(args []string)) {
	cn := GetCommandNode(mode)
	cn.commands = append(cn.commands, Command{m: match, f: f})
}

func GetCommandNodeCurrent() *CommandNode {
	return GetCommandNode(cliMode)
}

func listCallback(arg []string) {
	cn := GetCommandNodeCurrent()
	for _, cmd := range cn.commands {
		fmt.Printf("%s\n", cmd.m)
	}
}

func quitCallback(arg []string) {
	switch cliMode {
	case CliModeView:
		exit = true
	case CliModeConfigure:
		cliMode = CliModeView
	}
}
