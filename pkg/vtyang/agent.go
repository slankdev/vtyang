package vtyang

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/k0kubun/pp"
	"github.com/slankdev/vtyang/pkg/liner"
	"github.com/spf13/cobra"
)

func init() {
	logfile, err := os.OpenFile("/tmp/vtyang.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic("cannnot open test.log:" + err.Error())
	}
	log.SetOutput(logfile)
}

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
	line := liner.NewLiner()
	defer line.Close()
	line.SetCtrlCAborts(true)
	line.SetWordCompleter(completer)
	line.SetTabCompletionStyle(liner.TabPrints)
	line.SetBinder(63, binder) // 63 = Question Mark '?'

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
			log.Print("Aborted")
			break
		} else {
			log.Print("Error reading line: ", err)
		}
	}

	return nil
}
