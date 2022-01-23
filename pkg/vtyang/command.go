package vtyang

import (
	"fmt"
	"os"
	"sort"

	"github.com/openconfig/goyang/pkg/yang"
	"github.com/spf13/cobra"
)

var config = struct {
	GlobalOptDebug bool
	GlobalOptPaths []string
}{}

func NewCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:  "vtyang",
		RunE: main,
	}

	fs := rootCmd.PersistentFlags()
	fs.BoolVar(&config.GlobalOptDebug, "debug", false, "Enable debug output")
	fs.StringArrayVarP(&config.GlobalOptPaths, "path", "p", []string{}, "Module paths")
	rootCmd.AddCommand(newCommandCompletion(rootCmd))
	return rootCmd
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

func main(cmd *cobra.Command, args []string) error {
	files := []string{"./yang/data.yang"}
	ms := yang.NewModules()
	for _, name := range files {
		if err := ms.Read(name); err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
	}
	exitIfError(ms.Process())

	// Keep track of the top level modules we read in.
	// Those are the only modules we want to print below.
	mods := map[string]*yang.Module{}
	var names []string

	for _, m := range ms.Modules {
		if mods[m.Name] == nil {
			mods[m.Name] = m
			names = append(names, m.Name)
		}
	}
	sort.Strings(names)
	entries := make([]*yang.Entry, len(names))
	for x, n := range names {
		entries[x] = yang.ToEntry(mods[n])
	}

	//formatters["tree"].f(os.Stdout, entries)
	doTree(os.Stdout, entries)
	return nil
}
