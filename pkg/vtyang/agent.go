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
	QUESTION_MARK rune = 64
)

func completer(line string, pos int) (string, []string, string) {
	log.Printf("hoge")
	names := []string{"john", "james", "mary", "nancy"}
	return line[:pos], names, line[pos:]
}

func binder(line string) {
	fmt.Printf("\n\n")
	fmt.Printf("Possible Completions:\n")
	fmt.Printf(" hoge    sdkfjds sadf sdfsd fsdf\n")
	fmt.Printf(" fuga    sdkfjds sadf sdfsd fsdf\n")
	fmt.Printf("\n")
}

func agentMain(cmd *cobra.Command, args []string) error {
	m := NewDatabaseManager()
	if err := m.LoadYangModule("./yang"); err != nil {
		return err
	}

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
