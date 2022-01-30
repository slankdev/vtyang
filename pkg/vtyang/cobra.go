package vtyang

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

var config = struct {
	GlobalOptDebug       bool
	GlobalOptPaths       []string
	GlobalOptDBPath      string
	GlobalOptRunFilePath string
}{}

func NewCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "vtyang",
	}

	fs := rootCmd.PersistentFlags()
	fs.BoolVar(&config.GlobalOptDebug, "debug", false, "Enable debug output")
	fs.StringArrayVarP(&config.GlobalOptPaths, "path", "p", []string{}, "Module paths")
	fs.StringVarP(&config.GlobalOptDBPath, "dbpath", "d", "/etc/vtyang/config.json", "Database path")
	fs.StringVarP(&config.GlobalOptRunFilePath, "run-path", "r", "", "Runtime file path")
	rootCmd.AddCommand(newCommandCompletion(rootCmd))
	rootCmd.AddCommand(newCommandAgent())
	rootCmd.AddCommand(newCommandDump())
	return rootCmd
}

func newCommandAgent() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "agent",
		RunE: agentMain,
	}
	return cmd
}

func newCommandDump() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "dump",
		RunE: dumpMain,
	}
	return cmd
}

func newCommandCompletion(rootCmd *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [sub operation]",
		Short: "Display completion snippet",
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "bash",
		Short: "Display bash-completion snippet",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			rootCmd.GenBashCompletion(os.Stdout)
		},
		SilenceUsage: true,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "zsh",
		Short: "Display zsh-completion snippet",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			rootCmd.GenZshCompletion(os.Stdout)
		},
		SilenceUsage: true,
	})

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
