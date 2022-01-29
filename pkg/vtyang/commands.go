package vtyang

import (
	"fmt"
	"strings"
)

type Command struct {
	m string
	f func(args []string)
}

var commands []Command

func InstallCommands() {
	InstallCommand("write memory", func(args []string) {
		if err := dbm.db.root.WriteToJsonFile(config.GlobalOptDBPath); err != nil {
			fmt.Printf("Error: %s\n", err.Error())
		}
	})

	InstallCommand("show running-config", func(args []string) {
		fmt.Println(dbm.db.root.String())
	})

	InstallCommand("show operational-data", func(args []string) {
		if len(args) < 3 {
			fmt.Printf("usage:\n")
			return
		}

		xpath, _, err := ParseXPathArgs(args[2:], false)
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			return
		}

		node, err := dbm.GetNode(xpath)
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			return
		}
		if node == nil {
			fmt.Printf("Not Found\n")
			return
		}
		fmt.Println(node.String())
	})

	InstallCommand("set", func(args []string) {
		if len(args) < 2 {
			fmt.Printf("usage:\n")
			return
		}

		args = args[1:]
		xpath, valueStr, err := ParseXPathArgs(args, true)
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			return
		}

		if _, err := dbm.SetNode(xpath, valueStr); err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			return
		}
	})

	InstallCommand("delete", func(args []string) {
		if len(args) < 2 {
			fmt.Printf("usage:\n")
			return
		}

		args = args[1:]
		xpath, _, err := ParseXPathArgs(args, false)
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			return
		}

		if err := dbm.DeleteNode(xpath); err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			return
		}
	})
}

func InstallCommand(match string, f func(args []string)) {
	commands = append(commands, Command{m: match, f: f})
}

func ExecuteCommand(cli string) {
	args := strings.Fields(cli)
	notfound := true
	for _, cmd := range commands {
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
