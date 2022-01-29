package vtyang

import (
	"fmt"
	"log"
	"strings"

	"github.com/k0kubun/pp"
	"github.com/slankdev/vtyang/pkg/liner"
	"github.com/spf13/cobra"
)

const (
	// Question mark '?'
	QUESTION_MARK rune = 63
)

type CliMode int

const (
	CliModeView CliMode = iota
	CliModeConfigure
)

var cliMode CliMode = CliModeView
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

	setCompletionTreeForCommandShowOperationalData()
	setCompletionTreeForCommandSet()
	setCompletionTreeForCommandDelete()
	InstallCommands()
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
			case matchArgs(args, "show cli-tree"):
				pp.Println(tree)
			default:
				ExecuteCommand(cat(args))
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
