package vtyang

import (
	"log"
	"os"
	"strings"

	"github.com/k0kubun/pp"
	"github.com/slankdev/vtyang/pkg/liner"
	"github.com/spf13/cobra"
)

var (
	names = []string{"john", "james", "mary", "nancy"}
)

func completerA(line string) []string {
	log.Printf("hoge")

	c := []string{}
	for _, n := range names {
		if strings.HasPrefix(n, strings.ToLower(line)) {
			c = append(c, n)
		}
	}
	return c
}

func init() {
	logfile, err := os.OpenFile("/tmp/vtyang.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic("cannnot open test.log:" + err.Error())
	}
	log.SetOutput(logfile)
}

func agentMain(cmd *cobra.Command, args []string) error {
	line := liner.NewLiner()
	defer line.Close()
	line.SetCtrlCAborts(true)
	line.SetCompleter(completerA)
	line.SetTabCompletionStyle(liner.TabPrints)

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
