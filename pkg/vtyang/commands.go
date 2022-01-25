package vtyang

import (
	"fmt"
	"os"
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
	if node == nil {
		fmt.Printf("Not Found\n")
		return
	}

	node.Write(os.Stdout)
}
