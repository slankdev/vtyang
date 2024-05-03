package vtyang

import (
	"log"
	"os"

	"github.com/pkg/errors"
	"github.com/slankdev/vtyang/pkg/mgmtd"
)

type AgentOptsBackendMgmtd struct {
	UnixSockPath string
}

type AgentOpts struct {
	RuntimePath string
	YangPath    string
	LogFile     string
	// BackendMgmtd
	BackendMgmtd *AgentOptsBackendMgmtd
}

func InitAgent(opts AgentOpts) error {
	runtimePath := opts.RuntimePath
	yangPath := opts.YangPath
	logFile := opts.LogFile
	agentOpts = opts

	if logFile != "" {
		logfile, err := os.OpenFile(logFile,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			return errors.Wrapf(err, "os.OpenFile(%s)", logFile)
		}
		log.SetOutput(logfile)
		log.Printf("starting vtyang...\n")
	}

	if opts.BackendMgmtd != nil {
		var err error
		mgmtdClient, err = mgmtd.NewClient(
			opts.BackendMgmtd.UnixSockPath,
			"vtyang")
		if err != nil {
			return errors.Wrap(err, "mgmtd.NewClient")
		}
	}

	if runtimePath != "" {
		if err := os.MkdirAll(runtimePath, 0777); err != nil {
			return err
		}
	}

	GlobalOptRunFilePath = runtimePath
	dbm = NewDatabaseManager()
	if err := dbm.LoadDatabaseFromFile(getDatabasePath()); err != nil {
		return err
	}

	var err error
	yangmodules, err = yangModulesPath(yangPath)
	if err != nil {
		return err
	}

	cliMode = CliModeView
	commandnodes = nil
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
