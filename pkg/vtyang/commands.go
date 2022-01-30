package vtyang

import (
	"fmt"
)

func InstallCommands() {
	InstallCommand(CliModeView,
		"configure", []string{
			"Enable configure mode",
		},
		func(args []string) {
			if cliMode == CliModeConfigure {
				fmt.Printf("Already in configure mode\n")
				return
			}
			cliMode = CliModeConfigure
			dbm.candidateRoot = dbm.root.DeepCopy()
		})

	InstallCommand(CliModeConfigure,
		"show configuration diff", []string{
			"Display information",
			"Display configuration information",
			"Display configuration diff",
		},
		func(args []string) {
			diff := DBNodeDiff(&dbm.root, dbm.candidateRoot)
			fmt.Println(diff)
		})

	InstallCommand(CliModeConfigure,
		"commit", []string{
			"Commit current set of changes",
		}, ccbCommitCallback)

	InstallCommand(CliModeConfigure,
		"rollback configuration", []string{
			"Roll back database to last committed version",
			"Roll back database to last committed version",
		}, ccbRollbackConfiguration)

	InstallCommand(CliModeView,
		"write memory", []string{
			"Write system parameter",
			"Write system parameter to memory",
		},
		func(args []string) {
			if err := dbm.root.WriteToJsonFile(config.GlobalOptDBPath); err != nil {
				fmt.Printf("Error: %s\n", err.Error())
			}
		})

	InstallCommand(CliModeView,
		"show yang modules", []string{
			"Display information",
			"Display yang information",
			"Display yang modules",
		},
		func(args []string) {
			dbm.Dump()
		})

	InstallCommand(CliModeView,
		"show startup-config", []string{
			"Display information",
			"Display startup configuration",
		},
		func(args []string) {
			fmt.Println("not implemented")
		})

	InstallCommandYang(CliModeView,
		"show running-config", []string{
			"Display information",
			"Display current configuration",
		},
		func(args []string) {
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
			fmt.Println(node.String())
		}, yangCompletionForConfigurationData)

	InstallCommand(CliModeView,
		"show configuration commit list", []string{
			"Display information",
			"Display configuration",
			"Display configuration commit",
			"Display commit history",
		}, ccbShowConfigurationCommitList)

	InstallCommand(CliModeView,
		"show configuration commit diff", []string{
			"Display information",
			"Display configuration",
			"Display configuration commit",
			"Display configuration diff with history",
		}, ccbShowConfigurationCommitDiff)

	InstallCommandYang(CliModeView,
		"show", []string{
			"Display operational data",
		},
		func(args []string) {
			if len(args) < 2 {
				fmt.Printf("usage:\n")
				return
			}
			xpath, _, err := ParseXPathArgs(args[1:], false)
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
		}, yangCompletionForOperationalData)

	InstallCommandYang(CliModeConfigure,
		"no", []string{
			"Delete system parameter",
		},
		func(args []string) {
			xpath, _, err := ParseXPathArgs(args[1:], false)
			if err != nil {
				fmt.Printf("Error: %s\n", err.Error())
				return
			}
			if err := dbm.DeleteNode(xpath); err != nil {
				fmt.Printf("Error: %s\n", err.Error())
				return
			}
		}, yangCompletionForConfigurationData)

	InstallSpecialCommands()
}

func InstallSpecialCommands() {
	// Install set commands
	cn := GetCommandNode(CliModeConfigure)
	for _, e := range dbm.YangEntries() {
		cn.commands = append(cn.commands, Command{m: e.Name, f: func(args []string) {
			xpath, valueStr, err := ParseXPathArgs(args, true)
			if err != nil {
				fmt.Printf("Error: %s\n", err.Error())
				return
			}
			if _, err := dbm.SetNode(xpath, valueStr); err != nil {
				fmt.Printf("Error: %s\n", err.Error())
				return
			}
		}})
	}
	root := GetCommandNode(CliModeConfigure).tree.Root
	for _, e := range dbm.YangEntries() {
		n := yangCompletionForConfigurationData(e)
		if n != nil {
			root.Childs = append(root.Childs, n)
		}
	}

	// Install do command
	InstallCommand(CliModeConfigure, "do",
		[]string{"Run an operational-mode command"},
		func(args []string) {
			cn := GetCommandNode(CliModeView)
			cn.ExecuteCommand(cat(args[1:]))
		})
	viewRoot := GetCommandNode(CliModeView).tree.Root
	confRoot := GetCommandNode(CliModeConfigure).tree.Root
	for _, child := range confRoot.Childs {
		if child.Name == "do" {
			child.Childs = append(child.Childs, viewRoot.Childs...)
		}
	}
}
