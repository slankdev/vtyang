package vtyang

import (
	"fmt"
)

func InstallCommands() {
	InstallCommand(CliModeView,
		"configure",
		[]string{
			"Enable configure mode",
		},
		func(args []string) {
			cliMode = CliModeConfigure
		})

	InstallCommand(CliModeView,
		"write memory",
		[]string{
			"Write system parameter",
			"Write system parameter to memory",
		},
		func(args []string) {
			if err := dbm.db.root.WriteToJsonFile(config.GlobalOptDBPath); err != nil {
				fmt.Printf("Error: %s\n", err.Error())
			}
		})

	InstallCommand(CliModeView,
		"show running-config",
		[]string{
			"Display information",
			"Display current configuration",
		},
		func(args []string) {
			fmt.Println(dbm.db.root.String())
		})

	InstallCommand(CliModeView,
		"show operational-data",
		[]string{
			"Display information",
			"Display operational data",
		},
		func(args []string) {
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

	InstallCommand(CliModeConfigure,
		"set",
		[]string{
			"Set system parameter",
		},
		func(args []string) {
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

	InstallCommand(CliModeConfigure,
		"delete",
		[]string{
			"Delete system parameter",
		},
		func(args []string) {
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
