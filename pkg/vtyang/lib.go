package vtyang

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime/trace"
	"sort"
	"strings"

	"github.com/openconfig/goyang/pkg/indent"
	"github.com/openconfig/goyang/pkg/yang"
	"github.com/pborman/getopt"
)

var (
	S = "hoge"
)

func init() {
	print("hoge")
	register(&formatter{
		name: "tree",
		f:    doTree,
		help: "display in a tree format",
	})
}

func doTree(w io.Writer, entries []*yang.Entry) {
	for _, e := range entries {
		Write(w, e)
	}
}

// Write writes e, formatted, and all of its children, to w.
func Write(w io.Writer, e *yang.Entry) {
	if e.Description != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(indent.NewWriter(w, "// "), e.Description)
	}
	if len(e.Exts) > 0 {
		fmt.Fprintf(w, "extensions: {\n")
		for _, ext := range e.Exts {
			if n := ext.NName(); n != "" {
				fmt.Fprintf(w, "  %s %s;\n", ext.Kind(), n)
			} else {
				fmt.Fprintf(w, "  %s;\n", ext.Kind())
			}
		}
		fmt.Fprintln(w, "}")
	}
	switch {
	case e.RPC != nil:
		fmt.Fprintf(w, "RPC: ")
	case e.ReadOnly():
		fmt.Fprintf(w, "RO: ")
	default:
		fmt.Fprintf(w, "rw: ")
	}
	if e.Type != nil {
		fmt.Fprintf(w, "%s ", getTypeName(e))
	}
	name := e.Name
	if e.Prefix != nil {
		name = e.Prefix.Name + ":" + name
	}
	switch {
	case e.Dir == nil && e.ListAttr != nil:
		fmt.Fprintf(w, "[]%s\n", name)
		return
	case e.Dir == nil:
		fmt.Fprintf(w, "%s\n", name)
		return
	case e.ListAttr != nil:
		fmt.Fprintf(w, "[%s]%s {\n", e.Key, name) //}
	default:
		fmt.Fprintf(w, "%s {\n", name) //}
	}
	if r := e.RPC; r != nil {
		if r.Input != nil {
			Write(indent.NewWriter(w, "  "), r.Input)
		}
		if r.Output != nil {
			Write(indent.NewWriter(w, "  "), r.Output)
		}
	}
	var names []string
	for k := range e.Dir {
		names = append(names, k)
	}
	sort.Strings(names)
	for _, k := range names {
		Write(indent.NewWriter(w, "  "), e.Dir[k])
	}
	// { to match the brace below to keep brace matching working
	fmt.Fprintln(w, "}")
}

func getTypeName(e *yang.Entry) string {
	if e == nil || e.Type == nil {
		return ""
	}
	// Return our root's type name.
	// This is should be the builtin type-name
	// for this entry.
	return e.Type.Root.Name
}

var (
	typesDebug   bool
	typesVerbose bool
)

func init() {
	flags := getopt.New()
	register(&formatter{
		name:  "types",
		f:     doTypes,
		help:  "display found types",
		flags: flags,
	})
	flags.BoolVarLong(&typesDebug, "types_debug", 0, "display debug information")
	flags.BoolVarLong(&typesVerbose, "types_verbose", 0, "include base information")
}

func doTypes(w io.Writer, entries []*yang.Entry) {
	types := Types{}
	for _, e := range entries {
		types.AddEntry(e)
	}

	for t := range types {
		printType(w, t, typesVerbose)
	}
	if typesDebug {
		for _, e := range entries {
			showall(w, e)
		}
	}
}

// Types keeps track of all the YangTypes defined.
type Types map[*yang.YangType]struct{}

// AddEntry adds all types defined in e and its descendants to t.
func (t Types) AddEntry(e *yang.Entry) {
	if e == nil {
		return
	}
	if e.Type != nil {
		t[e.Type.Root] = struct{}{}
	}
	for _, d := range e.Dir {
		t.AddEntry(d)
	}
}

// printType prints type t in a moderately human readable format to w.
func printType(w io.Writer, t *yang.YangType, verbose bool) {
	if verbose && t.Base != nil {
		base := yang.Source(t.Base)
		if base == "unknown" {
			base = "unnamed type"
		}
		fmt.Fprintf(w, "%s: ", base)
	}
	fmt.Fprintf(w, "%s", t.Root.Name)
	if t.Kind.String() != t.Root.Name {
		fmt.Fprintf(w, "(%s)", t.Kind)
	}
	if t.Units != "" {
		fmt.Fprintf(w, " units=%s", t.Units)
	}
	if t.Default != "" {
		fmt.Fprintf(w, " default=%q", t.Default)
	}
	if t.FractionDigits != 0 {
		fmt.Fprintf(w, " fraction-digits=%d", t.FractionDigits)
	}
	if len(t.Length) > 0 {
		fmt.Fprintf(w, " length=%s", t.Length)
	}
	if t.Kind == yang.YinstanceIdentifier && !t.OptionalInstance {
		fmt.Fprintf(w, " required")
	}
	if t.Kind == yang.Yleafref && t.Path != "" {
		fmt.Fprintf(w, " path=%q", t.Path)
	}
	if len(t.Pattern) > 0 {
		fmt.Fprintf(w, " pattern=%s", strings.Join(t.Pattern, "|"))
	}
	b := yang.BaseTypedefs[t.Kind.String()].YangType
	if len(t.Range) > 0 && !t.Range.Equal(b.Range) {
		fmt.Fprintf(w, " range=%s", t.Range)
	}
	if len(t.Type) > 0 {
		fmt.Fprintf(w, "{\n")
		for _, t := range t.Type {
			printType(indent.NewWriter(w, "  "), t, verbose)
		}
		fmt.Fprintf(w, "}")
	}
	fmt.Fprintf(w, ";\n")
}

func showall(w io.Writer, e *yang.Entry) {
	if e == nil {
		return
	}
	if e.Type != nil {
		fmt.Fprintf(w, "\n%s\n  ", e.Node.Statement().Location())
		printType(w, e.Type.Root, false)
	}
	for _, d := range e.Dir {
		showall(w, d)
	}
}

// Each format must register a formatter with register.  The function f will
// be called once with the set of yang Entry trees generated.
type formatter struct {
	name  string
	f     func(io.Writer, []*yang.Entry)
	help  string
	flags *getopt.Set
}

var formatters = map[string]*formatter{}

func register(f *formatter) {
	formatters[f.name] = f
}

// exitIfError writes errs to standard error and exits with an exit status of 1.
// If errs is empty then exitIfError does nothing and simply returns.
func exitIfError(errs []error) {
	if len(errs) > 0 {
		for _, err := range errs {
			fmt.Fprintln(os.Stderr, err)
		}
		stop(1)
	}
}

var stop = os.Exit

func main() {
	var format string
	formats := make([]string, 0, len(formatters))
	for k := range formatters {
		formats = append(formats, k)
	}
	sort.Strings(formats)

	var traceP string
	var help bool
	var paths []string
	getopt.ListVarLong(&paths, "path", 'p', "comma separated list of directories to add to search path", "DIR[,DIR...]")
	getopt.StringVarLong(&format, "format", 'f', "format to display: "+strings.Join(formats, ", "), "FORMAT")
	getopt.StringVarLong(&traceP, "trace", 't', "write trace into to TRACEFILE", "TRACEFILE")
	getopt.BoolVarLong(&help, "help", 'h', "display help")
	getopt.BoolVarLong(&yang.ParseOptions.IgnoreSubmoduleCircularDependencies, "ignore-circdep", 'g', "ignore circular dependencies between submodules")
	getopt.SetParameters("[FORMAT OPTIONS] [SOURCE] [...]")

	if err := getopt.Getopt(func(o getopt.Option) bool {
		if o.Name() == "--format" {
			f, ok := formatters[format]
			if !ok {
				fmt.Fprintf(os.Stderr, "%s: invalid format.  Choices are %s\n", format, strings.Join(formats, ", "))
				stop(1)
			}
			if f.flags != nil {
				f.flags.VisitAll(func(o getopt.Option) {
					getopt.AddOption(o)
				})
			}
		}
		return true
	}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		getopt.PrintUsage(os.Stderr)
		os.Exit(1)
	}

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
		getopt.CommandLine.PrintUsage(os.Stderr)
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

	files := getopt.Args()

	ms := yang.NewModules()

	if len(files) == 0 {
		data, err := ioutil.ReadAll(os.Stdin)
		if err == nil {
			err = ms.Parse(string(data), "<STDIN>")
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			stop(1)
		}
	}

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
}
