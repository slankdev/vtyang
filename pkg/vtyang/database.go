package vtyang

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

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

	entries := []*yang.Entry{}
	for _, m := range modules.Modules {
		entries = append(entries, yang.ToEntry(m))
	}

	for _, e := range entries {
		write(os.Stdout, e)
	}
	return nil
}
