package vtyang

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/openconfig/goyang/pkg/yang"
	"github.com/slankdev/vtyang/pkg/liner"
	"github.com/slankdev/vtyang/pkg/util"
	"github.com/spf13/cobra"
)

var (
	GlobalOptRunFilePath string

	actionCBs = map[string]func(args []string) error{
		"uptime-callback": func(args []string) error {
			fmt.Fprintf(stdout, "UPTIME")
			return nil
		},
		"date-callback": func(args []string) error {
			fmt.Fprint(stdout, "DATE")
			return nil
		},
	}
	_ = actionCBs

	exit            bool      = false
	stdout          io.Writer = os.Stdout
	cliMode         CliMode   = CliModeView
	dbm             *DatabaseManager
	commitHistories []CommitHistory
	commandnodes    map[CliMode]*CommandNode
	yangmodules     *yang.Modules
)

const (
	QUESTION_MARK rune = 63 // '?'
)

func getDatabasePath() string {
	return fmt.Sprintf("%s/config.json", GlobalOptRunFilePath)
}

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

func NewCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "vtyang",
	}
	rootCmd.AddCommand(util.NewCommandCompletion(rootCmd))
	rootCmd.AddCommand(newCommandAgent())
	return rootCmd
}

func InitAgent(runtimePath, yangPath string) error {
	if runtimePath != "" {
		if err := os.MkdirAll(runtimePath, 0777); err != nil {
			return err
		}
	}

	GlobalOptRunFilePath = runtimePath
	dbm = NewDatabaseManager()
	dbm.LoadYangModuleOrDie(yangPath)
	if err := dbm.LoadDatabaseFromFile(getDatabasePath()); err != nil {
		return err
	}

	var err error
	yangmodules, err = yangModulesPath(yangPath)
	if err != nil {
		return err
	}

	cliMode = CliModeView
	installCommandsDefault(CliModeView)
	installCommandsDefault(CliModeConfigure)
	installCommands()
	initCommitHistories()
	installCommandsPostProcess()

	if GlobalOptRunFilePath != "" {
		if err := os.MkdirAll(GlobalOptRunFilePath, 0777); err != nil {
			return err
		}
	}
	return nil
}

func newCommandAgent() *cobra.Command {
	cmd := &cobra.Command{
		Use: "agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			// XXX
			if err := InitAgent(GlobalOptRunFilePath, "./yang"); err != nil {
				return err
			}

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
					cn := getCommandNodeCurrent()
					cn.executeCommand(cat(args))
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
		},
	}
	fs := cmd.Flags()
	fs.StringVarP(&GlobalOptRunFilePath, "run-path", "r", "", "Runtime file path")
	return cmd
}

func init() {
	logfile, err := os.OpenFile("/tmp/vtyang.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic("cannnot open test.log:" + err.Error())
	}
	log.SetOutput(logfile)
	log.Printf("starting vtyang...\n")
}

func ErrorOnDie(err error) {
	if err != nil {
		panic(err)
	}
}
