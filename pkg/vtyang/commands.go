package vtyang

import (
	"fmt"
)

func InstallCommands() {
	InstallCommand("show running-config", func(args []string) {
		fmt.Println(dbm.db.root.String())
	})

	InstallCommand("show operational-data", func(args []string) {
		if len(args) < 3 {
			fmt.Printf("usage:\n")
			return
		}

		mod, xpath, _, err := ParseXPathArgs(args[2:], false)
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			return
		}

		node, err := dbm.GetNode(mod, xpath)
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
		mod, xpath, valueStr, err := ParseXPathArgs(args, true)
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			return
		}

		if _, err := dbm.SetNode(mod, xpath, valueStr); err != nil {
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
		mod, xpath, _, err := ParseXPathArgs(args, false)
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			return
		}

		if err := dbm.DeleteNode(mod, xpath); err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			return
		}
	})
}
