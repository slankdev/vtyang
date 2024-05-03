package vtyang

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/openconfig/goyang/pkg/yang"
	"github.com/spf13/cobra"

	"github.com/slankdev/vtyang/pkg/liner"
	"github.com/slankdev/vtyang/pkg/mgmtd"
	"github.com/slankdev/vtyang/pkg/util"
)

var (
	GlobalOptEnableGrpc  bool
	GlobalOptLogFile     string
	GlobalOptRunFilePath string
	GlobalOptYangPath    []string
	GlobalOptDumpCliTree string
	GlobalOptCommands    []string
	GlobalOptMgmtdSock   string

	agentOpts   AgentOpts
	mgmtdClient *mgmtd.Client

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
		RunE: func(cmd *cobra.Command, args []string) error {
			// Prepare agent object
			opts := AgentOpts{
				RuntimePath: GlobalOptRunFilePath,
				YangPath:    GlobalOptYangPath,
				LogFile:     GlobalOptLogFile,
			}
			if GlobalOptMgmtdSock != "" {
				opts.BackendMgmtd = &AgentOptsBackendMgmtd{
					UnixSockPath: GlobalOptMgmtdSock,
				}
			}
			if err := InitAgent(opts); err != nil {
				return err
			}

			// Execute Commands as inline mode
			if len(GlobalOptCommands) > 0 {
				for _, c := range GlobalOptCommands {
					getCommandNodeCurrent().executeCommand(c)
				}
			}

			// Prepare shell objects
			line := liner.NewLiner()
			defer line.Close()
			line.SetCtrlCAborts(true)
			line.SetWordCompleter(completer)
			line.SetTabCompletionStyle(liner.TabPrints)
			line.SetBinder(QUESTION_MARK, completionLister)

			// Start shell loop
			for !exit {
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
			}
			return nil
		},
	}

	fs := rootCmd.Flags()
	fs.BoolVar(&GlobalOptEnableGrpc, "grpc", false, "Enable gRPC server")
	fs.StringVarP(&GlobalOptLogFile, "logfile", "l", "/tmp/vtyang.log", "Log file")
	fs.StringVarP(&GlobalOptRunFilePath, "run", "r", "", "Runtime file path")
	fs.StringArrayVarP(&GlobalOptYangPath, "yang", "y", []string{}, "Yang file path")
	fs.StringArrayVarP(&GlobalOptCommands, "command", "c", []string{}, "")
	fs.StringVar(&GlobalOptMgmtdSock, "mgmtd-sock", "", "/var/run/frr/mgmtd_fe.sock")

	rootCmd.AddCommand(util.NewCommandCompletion(rootCmd))
	rootCmd.AddCommand(util.NewCommandVersion())
	return rootCmd
}
