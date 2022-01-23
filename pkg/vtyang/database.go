package vtyang

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"

	"github.com/openconfig/goyang/pkg/yang"
)

type DatabaseManager struct {
}

func NewDatabaseManager() *DatabaseManager {
	m := DatabaseManager{}
	return &m
}

func (m *DatabaseManager) LoadYangModule(path string) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	modules := yang.NewModules()
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fullname := fmt.Sprintf("%s/%s", path, file.Name())
		log.Printf("loading yang module '%s'\n", fullname)
		if err := modules.Read(fullname); err != nil {
			return err
		}
	}

	errs := modules.Process()
	if len(errs) > 0 {
		for _, err := range errs {
			log.Printf("%s", err.Error())
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
