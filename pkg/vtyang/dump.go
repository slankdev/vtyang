package vtyang

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/openconfig/goyang/pkg/indent"
	"github.com/openconfig/goyang/pkg/yang"
	"github.com/spf13/cobra"
)

func write(w io.Writer, e *yang.Entry) {
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
			write(indent.NewWriter(w, "  "), r.Input)
		}
		if r.Output != nil {
			write(indent.NewWriter(w, "  "), r.Output)
		}
	}
	var names []string
	for k := range e.Dir {
		names = append(names, k)
	}
	sort.Strings(names)
	for _, k := range names {
		write(indent.NewWriter(w, "  "), e.Dir[k])
	}
	// { to match the brace below to keep brace matching working
	fmt.Fprintln(w, "}")
}

func getTypeName(e *yang.Entry) string {
	if e == nil || e.Type == nil {
		return ""
	}
	return e.Type.Root.Name
}

func dumpMain(cmd *cobra.Command, args []string) error {
	files := []string{
		"./yang/model1.yang",
		"./yang/model2.yang",
	}
	modules := yang.NewModules()
	for _, name := range files {
		if err := modules.Read(name); err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
	}

	errs := modules.Process()
	if len(errs) > 0 {
		for _, err := range errs {
			fmt.Fprintln(os.Stderr, err)
		}
		return errs[0]
	}

	mods := map[string]*yang.Module{}
	var names []string

	for _, m := range modules.Modules {
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

	for _, e := range entries {
		write(os.Stdout, e)
	}
	return nil
}
