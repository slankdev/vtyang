package vtyang

import (
	"fmt"
	"log"
	"strings"

	"github.com/k0kubun/pp"
	"github.com/slankdev/vtyang/pkg/liner"
	"github.com/spf13/cobra"
)

const (
	// Question mark '?'
	QUESTION_MARK rune = 63
)

var dbm *DatabaseManager

func completer(line string, pos int) (string, []string, string) {
	log.Printf("hoge")
	names := []string{"john", "james", "mary", "nancy"}
	return line[:pos], names, line[pos:]
}

func binder(line string, pos int) {
	fmt.Printf("\n")
	fmt.Printf("Possible Completions:\n")
	ents := dbm.DumpEntries()
	for _, ent := range ents {
		fmt.Printf("  %s  %s\n", ent.Name, ent.Description)
	}
}

func agentMain(cmd *cobra.Command, args []string) error {
	dbm = NewDatabaseManager()
	dbm.LoadYangModuleOrDie("./yang")
	dbm.Dump()

	line := liner.NewLiner()
	defer line.Close()
	line.SetCtrlCAborts(true)
	line.SetWordCompleter(completer)
	line.SetTabCompletionStyle(liner.TabPrints)
	line.SetBinder(QUESTION_MARK, binder)

	for {
		if name, err := line.Prompt("vtyang# "); err == nil {
			line.AppendHistory(name)
			name = strings.TrimSpace(name)
			args := strings.Fields(name)
			if len(args) == 0 {
				continue

			}

			switch args[0] {
			default:
				pp.Println(args)
			}

		} else if err == liner.ErrPromptAborted {
			log.Print("aborted")
			break
		} else {
			log.Print("error reading line: ", err)
		}
	}

	return nil
}
