package vtyang

import (
	"fmt"
	"log"
	"strings"

	"github.com/slankdev/vtyang/pkg/liner"
	"github.com/spf13/cobra"
)

const (
	// Question mark '?'
	QUESTION_MARK rune = 63
)

var exit bool = false

func getPrompt() string {
	switch cliMode {
	case CliModeView:
		return "vtyang# "
	case CliModeConfigure:
		return "vtyang(config)# "
	default:
		panic(fmt.Sprintf("CLIMODE(%v)", cliMode))
	}
}

func agentMain(cmd *cobra.Command, args []string) error {
	dbm = NewDatabaseManager()
	dbm.LoadYangModuleOrDie("./yang")
	if err := dbm.LoadDatabaseFromFile(config.GlobalOptDBPath); err != nil {
		return err
	}

	InstallCommands()
	InstallCommandsDefault(CliModeView)
	InstallCommandsDefault(CliModeConfigure)
	setCompletionTreeForCommandShowOperationalData()
	setCompletionTreeForCommandSet()
	setCompletionTreeForCommandDelete()
	InitVTYang()

	line := liner.NewLiner()
	defer line.Close()
	line.SetCtrlCAborts(true)
	line.SetWordCompleter(completer)
	line.SetTabCompletionStyle(liner.TabPrints)
	line.SetBinder(QUESTION_MARK, completionLister)

	for {
		if name, err := line.Prompt(getPrompt()); err == nil {
			line.AppendHistory(name)
			name = strings.TrimSpace(name)
			args := strings.Fields(name)
			if len(args) == 0 {
				continue
			}

			switch {
			case matchArgs(args, "show yang-modules"):
				dbm.Dump()
			default:
				cn := GetCommandNodeCurrent()
				cn.ExecuteCommand(cat(args))
			}

		} else if err == liner.ErrPromptAborted {
			log.Print("aborted")
			break
		} else {
			log.Print("error reading line: ", err)
		}

		if exit {
			break
		}
	}

	return nil
}
