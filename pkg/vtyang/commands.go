package vtyang

import (
	"fmt"
)

func InstallCommands() {
	InstallCommand(CliModeView, "configure", func(args []string) {
		cliMode = CliModeConfigure
	})

	quit := func(args []string) {
		switch cliMode {
		case CliModeView:
			exit = true
		case CliModeConfigure:
			cliMode = CliModeView
		}
	}
	InstallCommand(CliModeView, "quit", quit)
	InstallCommand(CliModeConfigure, "quit", quit)

	InstallCommand(CliModeView, "write memory", func(args []string) {
		if err := dbm.db.root.WriteToJsonFile(config.GlobalOptDBPath); err != nil {
			fmt.Printf("Error: %s\n", err.Error())
		}
	})

	InstallCommand(CliModeView, "show running-config", func(args []string) {
		fmt.Println(dbm.db.root.String())
	})

	InstallCommand(CliModeView, "show operational-data", func(args []string) {
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

	InstallCommand(CliModeConfigure, "set", func(args []string) {
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

	InstallCommand(CliModeConfigure, "delete", func(args []string) {
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
