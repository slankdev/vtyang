package vtyang

import (
	"fmt"
	"os"
	"runtime/trace"
	"sort"
	"strings"

	"github.com/openconfig/goyang/pkg/indent"
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
		RunE: f,
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

func f(cmd *cobra.Command, args []string) error {
	var format string
	formats := make([]string, 0, len(formatters))
	for k := range formatters {
		formats = append(formats, k)
	}
	sort.Strings(formats)

	var traceP string
	var help bool
	var paths []string
	// getopt.ListVarLong(&paths, "path", 'p', "comma separated list of directories to add to search path", "DIR[,DIR...]")
	// getopt.StringVarLong(&format, "format", 'f', "format to display: "+strings.Join(formats, ", "), "FORMAT")
	// getopt.StringVarLong(&traceP, "trace", 't', "write trace into to TRACEFILE", "TRACEFILE")
	// getopt.BoolVarLong(&help, "help", 'h', "display help")
	// getopt.BoolVarLong(&yang.ParseOptions.IgnoreSubmoduleCircularDependencies, "ignore-circdep", 'g', "ignore circular dependencies between submodules")
	// getopt.SetParameters("[FORMAT OPTIONS] [SOURCE] [...]")
	// if err := getopt.Getopt(func(o getopt.Option) bool {
	// 	if o.Name() == "--format" {
	// 		f, ok := formatters[format]
	// 		if !ok {
	// 			fmt.Fprintf(os.Stderr, "%s: invalid format.  Choices are %s\n", format, strings.Join(formats, ", "))
	// 			stop(1)
	// 		}
	// 		if f.flags != nil {
	// 			f.flags.VisitAll(func(o getopt.Option) {
	// 				getopt.AddOption(o)
	// 			})
	// 		}
	// 	}
	// 	return true
	// }); err != nil {
	// 	fmt.Fprintln(os.Stderr, err)
	// 	getopt.PrintUsage(os.Stderr)
	// 	os.Exit(1)
	// }

	if traceP != "" {
		fp, err := os.Create(traceP)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		trace.Start(fp)
		stop = func(c int) { trace.Stop(); os.Exit(c) }
		defer func() { trace.Stop() }()
	}

	if help {
		//getopt.CommandLine.PrintUsage(os.Stderr)
		fmt.Fprintf(os.Stderr, `
SOURCE may be a module name or a .yang file.
Formats:
`)
		for _, fn := range formats {
			f := formatters[fn]
			fmt.Fprintf(os.Stderr, "    %s - %s\n", f.name, f.help)
			if f.flags != nil {
				f.flags.PrintOptions(indent.NewWriter(os.Stderr, "   "))
			}
			fmt.Fprintln(os.Stderr)
		}
		stop(0)
	}

	for _, path := range paths {
		expanded, err := yang.PathsWithModules(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		yang.AddPath(expanded...)
	}

	if format == "" {
		format = "tree"
	}
	if _, ok := formatters[format]; !ok {
		fmt.Fprintf(os.Stderr, "%s: invalid format.  Choices are %s\n", format, strings.Join(formats, ", "))
		stop(1)

	}

	files := []string{
		"./yang/data.yang",
	}

	ms := yang.NewModules()
	for _, name := range files {
		if err := ms.Read(name); err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
	}

	// Process the read files, exiting if any errors were found.
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

	formatters[format].f(os.Stdout, entries)
	return nil
}
