package vtyang

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sort"

	"github.com/openconfig/goyang/pkg/indent"
	"github.com/openconfig/goyang/pkg/yang"
)

type DatabaseManager struct {
	modules *yang.Modules
}

var dbm *DatabaseManager

func NewDatabaseManager() *DatabaseManager {
	m := DatabaseManager{}
	m.modules = yang.NewModules()
	return &m
}

func (m *DatabaseManager) LoadYangModuleOrDie(path string) {
	if err := m.LoadYangModule("./yang"); err != nil {
		panic(err)
	}
}

func (m *DatabaseManager) LoadYangModule(path string) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fullname := fmt.Sprintf("%s/%s", path, file.Name())
		log.Printf("loading yang module '%s'\n", fullname)
		if err := m.modules.Read(fullname); err != nil {
			return err
		}
	}

	errs := m.modules.Process()
	if len(errs) > 0 {
		for _, err := range errs {
			log.Printf("%s", err.Error())
		}
		return errs[0]
	}

	return nil
}

func (m *DatabaseManager) DumpEntries() []*yang.Entry {
	entries := []*yang.Entry{}
	for _, m := range m.modules.Modules {
		entries = append(entries, yang.ToEntry(m))
	}
	return entries
}

func (m *DatabaseManager) Dump() {
	entries := m.DumpEntries()
	for _, e := range entries {
		dump(os.Stdout, e)
	}
}

func dump(w io.Writer, e *yang.Entry) {
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

	if e.RPC != nil {
		return
	}

	switch {

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
			dump(indent.NewWriter(w, "  "), r.Input)
		}
		if r.Output != nil {
			dump(indent.NewWriter(w, "  "), r.Output)
		}
	}
	var names []string
	for k := range e.Dir {
		names = append(names, k)
	}
	sort.Strings(names)
	for _, k := range names {
		dump(indent.NewWriter(w, "  "), e.Dir[k])
	}
	// { to match the brace below to keep brace matching working
	fmt.Fprintln(w, "}")
}
