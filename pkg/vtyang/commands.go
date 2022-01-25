package vtyang

import (
	"fmt"
	"os"
)

func InstallCommands() {
	InstallCommand("show operational-data", func(args []string) {
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
		if node == nil {
			fmt.Printf("Not Found\n")
			return
		}
		node.Write(os.Stdout)
	})

	InstallCommand("set", func(args []string) {
		if len(args) < 3 {
			fmt.Printf("usage:\n")
			return
		}

		valueStr := args[len(args)-1]
		args = args[1 : len(args)-1]
		// pp.Println(valueStr)
		// pp.Println(args)

		mod, xpath := dbm.CraftXPath(args)
		// fmt.Printf("DEBUG xpath=%s\n", xpath)

		node, err := dbm.GetNode(mod, xpath)
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			return
		}
		if node == nil {
			fmt.Printf("Error: Resource Not Found\n")
			return
		}
		if node.Type != Leaf {
			fmt.Printf("Error: Specified path is not Leaf\n")
			return
		}

		node.Value.SetFromString(valueStr)
		//node.Write(os.Stdout)
	})
}
