package vtyang

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

type CommitHistory struct {
	Timestamp time.Time
	Before    string
	After     string
	Client    string
	Comment   string
}

var commitHistories []CommitHistory

func initCommitHistories() {
	if config.GlobalOptRunFilePath != "" {
		files, err := ioutil.ReadDir(config.GlobalOptRunFilePath)
		if err != nil {
			panic(err)
		}
		for _, file := range files {
			if strings.HasPrefix(file.Name(), "history") {
				fn := config.GlobalOptRunFilePath + "/" + file.Name()
				h, err := ReadCommitHistoryFromFile(fn)
				if err != nil {
					panic(err)
				}
				commitHistories = append([]CommitHistory{h}, commitHistories...)
			}
		}
	}
}

func ReadCommitHistoryFromFile(filename string) (CommitHistory, error) {
	h := CommitHistory{}
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return h, err
	}

	m := map[string]interface{}{}
	if err := json.Unmarshal(b, &m); err != nil {
		return h, err
	}

	nb, err := Interface2DBNode(m["before"])
	if err != nil {
		return h, err
	}
	na, err := Interface2DBNode(m["after"])
	if err != nil {
		return h, err
	}

	h.Timestamp = time.Unix(0, int64(m["timestamp"].(float64)))
	h.Client = m["client"].(string)
	h.Comment = m["comment"].(string)
	h.Before = nb.String()
	h.After = na.String()

	return h, nil
}

func (h CommitHistory) ToDBNode() (*DBNode, error) {
	n, err := ReadFromJsonString(h.After)
	return n, err
}

func (h CommitHistory) WriteToFile(basepath string) error {
	bs, err := jsonstring2map(h.Before)
	if err != nil {
		return err
	}
	as, err := jsonstring2map(h.After)
	if err != nil {
		return err
	}

	m := map[string]interface{}{}
	m["timestamp"] = h.Timestamp.UnixNano()
	m["client"] = h.Client
	m["comment"] = h.Comment
	m["before"] = bs
	m["after"] = as

	b, err := json.Marshal(&m)
	if err != nil {
		return err
	}
	filename := fmt.Sprintf("%s/history.%d.json", basepath, h.Timestamp.UnixNano())
	if err := ioutil.WriteFile(filename, b, 0644); err != nil {
		return err
	}
	return nil
}

func ccbCommitCallback(args []string) {
	if dbm.candidateRoot == nil {
		panic("OKASHII")
	}

	h := CommitHistory{
		Before:    dbm.root.String(),
		After:     dbm.candidateRoot.String(),
		Client:    "cli",
		Timestamp: time.Now(),
	}
	if h.Comment == "" {
		h.Comment = "-"
	}
	commitHistories = append([]CommitHistory{h}, commitHistories...)
	if config.GlobalOptRunFilePath != "" {
		if err := h.WriteToFile(config.GlobalOptRunFilePath); err != nil {
			fmt.Printf("Warning: %s ... ignored\n", err.Error())
		}
	}

	cliMode = CliModeConfigure
	dbm.root = *dbm.candidateRoot.DeepCopy()
	if err := dbm.root.WriteToJsonFile(config.GlobalOptDBPath); err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
}

func ccbShowConfigurationCommitList(args []string) {
	if len(args) > 4 {
		idx, err := strconv.Atoi(args[4])
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			return
		}
		if len(commitHistories) < idx {
			fmt.Printf("Error: Invalid commit idx\n")
			return
		}
		history := commitHistories[idx]
		na, err1 := ReadFromJsonString(history.Before)
		nb, err2 := ReadFromJsonString(history.After)
		if err1 != nil || err2 != nil {
			fmt.Printf("Error1: %s\n", err1.Error())
			fmt.Printf("Error2: %s\n", err2.Error())
			return
		}
		diff := DBNodeDiff(na, nb)
		fmt.Println(diff)

	} else {
		table := newTable()
		table.SetHeader([]string{"Idx", "ID", "Timestamp", "Client", "Comment"})
		for idx, history := range commitHistories {
			table.Append([]string{
				strconv.Itoa(idx),
				strconv.FormatInt(history.Timestamp.UnixNano(), 10),
				history.Timestamp.Format("2006-01-02 15:04:05"),
				history.Client,
				history.Comment,
			})
		}
		table.Render()
	}
}

func ccbShowConfigurationCommitDiff(args []string) {
	if len(args) < 4 {
		fmt.Printf("Usage\n")
		return
	}

	idx, err := strconv.Atoi(args[4])
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return
	}
	if len(commitHistories) < idx {
		fmt.Printf("Error: Invalid commit idx\n")
		return
	}

	history := commitHistories[idx]
	node, err := history.ToDBNode()
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return
	}

	fmt.Println(DBNodeDiff(&dbm.root, node))
}

func ccbRollbackConfiguration(args []string) {
	if len(args) < 3 {
		fmt.Printf("Usage: rollback configuration <idx>\n")
		return
	}
	idxArg := args[2]

	idx, err := strconv.Atoi(idxArg)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return
	}
	if len(commitHistories) < idx {
		fmt.Printf("Error: Invalid commit idx\n")
		return
	}

	history := commitHistories[idx]
	node, err := history.ToDBNode()
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return
	}

	dbm.candidateRoot = node.DeepCopy()
}
