package vtyang

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/openconfig/goyang/pkg/yang"
)

type CliMode int

const (
	CliModeView CliMode = iota
	CliModeConfigure
)

type Command struct {
	m string
	f func(args []string)
}

type CommandNode struct {
	mode     CliMode
	commands []Command
	tree     CompletionTree
}

func getCommandNode(mode CliMode) *CommandNode {
	if commandnodes == nil {
		commandnodes = map[CliMode]*CommandNode{}
	}

	cn, ok := commandnodes[mode]
	if !ok {
		ncn := CommandNode{}
		ncn.mode = mode
		ncn.tree = CompletionTree{}
		ncn.tree.Root = &CompletionNode{}
		commandnodes[mode] = &ncn
		cn = commandnodes[mode]
	}
	return cn
}

func getCommandNodeCurrent() *CommandNode {
	return getCommandNode(cliMode)
}

func (cn *CommandNode) executeCommand(cli string) {
	args := strings.Fields(cli)
	notfound := true
	for _, cmd := range cn.commands {
		if matchArgs(args, cmd.m) {
			cmd.f(args)
			notfound = false
			break
		}
	}
	if notfound {
		fmt.Fprintf(stdout, "Error: command %s not found\n", args[0])
	}
}

func installCommandsDefault(mode CliMode) {
	installCommand(mode, "list", []string{"List cli nodes"}, func(arg []string) {
		cn := getCommandNodeCurrent()
		for _, cmd := range cn.commands {
			fmt.Fprintf(stdout, "%s\n", cmd.m)
		}
	})

	installCommand(mode, "quit", []string{"Quit system"}, func(arg []string) {
		switch cliMode {
		case CliModeView:
			exit = true
		case CliModeConfigure:
			cliMode = CliModeView
			dbm.candidateRoot = nil
		}
	})

	installCommand(mode, "show cli-tree", []string{
		"Display information",
		"Display completion tree",
	}, func(arg []string) {
		fmt.Fprintln(stdout, dumpCompletionTreeJson(getCommandNodeCurrent().tree.Root))
	})
}

func installCommand(mode CliMode, match string, helps []string,
	f func(args []string)) {
	cn := getCommandNode(mode)
	cn.commands = append(cn.commands, Command{m: match, f: f})

	args := strings.Fields(match)
	if len(args) != len(helps) {
		panic(fmt.Sprintf("ERROR %s len(helps)=%d", match, len(helps)))
	}

	node := cn.tree.Root
	var nn *CompletionNode
	for ; len(args) != 0; args, helps = args[1:], helps[1:] {
		for idx := range node.Childs {
			child := node.Childs[idx]
			if child.Name == args[0] {
				node = child
				goto end
			}
		}

		nn = new(CompletionNode)
		nn.Name = args[0]
		nn.Description = helps[0]
		if len(args) == 1 {
			nn.Childs = []*CompletionNode{{Name: "<cr>"}}
		}
		node.Childs = append(node.Childs, nn)
		node = node.Childs[len(node.Childs)-1]
	end:
	}
}

func installCommandsPostProcess() {
	cn0 := getCommandNode(CliModeView)
	sort.Slice(cn0.commands, func(i, j int) bool {
		return len(cn0.commands[i].m) > len(cn0.commands[j].m)
	})
	cn1 := getCommandNode(CliModeConfigure)
	sort.Slice(cn1.commands, func(i, j int) bool {
		return len(cn1.commands[i].m) > len(cn1.commands[j].m)
	})
}

func installCommands() {
	installCommand(CliModeView,
		"configure", []string{
			"Enable configure mode",
		},
		func(args []string) {
			if cliMode == CliModeConfigure {
				fmt.Fprintf(stdout, "Already in configure mode\n")
				return
			}
			cliMode = CliModeConfigure
			dbm.candidateRoot = dbm.root.DeepCopy()
		})

	installCommand(CliModeConfigure,
		"show configuration diff", []string{
			"Display information",
			"Display configuration information",
			"Display configuration diff",
		},
		func(args []string) {
			diff := DBNodeDiff(&dbm.root, dbm.candidateRoot)
			fmt.Fprintln(stdout, diff)
		})

	installCommand(CliModeConfigure,
		"commit", []string{
			"Commit current set of changes",
		}, ccbCommitCallback)

	installCommand(CliModeConfigure,
		"rollback configuration", []string{
			"Roll back database to last committed version",
			"Roll back database to last committed version",
		}, ccbRollbackConfiguration)

	installCommand(CliModeView,
		"write memory", []string{
			"Write system parameter",
			"Write system parameter to memory",
		},
		func(args []string) {
			if err := dbm.root.WriteToJsonFile(getDatabasePath()); err != nil {
				fmt.Fprintf(stdout, "Error: %s\n", err.Error())
			}
		})

	installCommand(CliModeView,
		"show yang modules", []string{
			"Display information",
			"Display yang information",
			"Display yang modules",
		},
		func(args []string) {
			dbm.Dump()
		})

	installCommand(CliModeView,
		"show startup-config", []string{
			"Display information",
			"Display startup configuration",
		},
		func(args []string) {
			fmt.Fprintln(stdout, "not implemented")
		})

	installCommand(CliModeView,
		"show configuration commit list", []string{
			"Display information",
			"Display configuration",
			"Display configuration commit",
			"Display commit history",
		}, ccbShowConfigurationCommitList)

	installCommand(CliModeView,
		"show configuration commit diff", []string{
			"Display information",
			"Display configuration",
			"Display configuration commit",
			"Display configuration diff with history",
		}, ccbShowConfigurationCommitDiff)

	installCommand(CliModeConfigure, "do",
		[]string{"Run an operational-mode command"},
		func(args []string) {
			cn := getCommandNode(CliModeView)
			cn.executeCommand(cat(args[1:]))
		})
	viewRoot := getCommandNode(CliModeView).tree.Root
	confRoot := getCommandNode(CliModeConfigure).tree.Root
	for _, child := range confRoot.Childs {
		if child.Name == "do" {
			child.Childs = append(child.Childs, viewRoot.Childs...)
		}
	}

	// Load Yang modules
	cn0 := getCommandNode(CliModeView)
	cn1 := getCommandNode(CliModeConfigure)

	// CLI CALLBAC
	// Configure mode
	cn0.commands = append(cn0.commands, getCommandCallbackRPC(yangmodules)...)
	cn0.commands = append(cn0.commands, getCommandCallbackOperState(yangmodules)...)
	cn0.commands = append(cn0.commands, getViewCommandCallbackConfig(yangmodules)...)
	cn1.commands = append(cn1.commands, getCommandCallbackConfig(yangmodules)...)

	// COMPLETION TREE
	// Normal,Configure mode
	installCompletionTree(cn0.tree.Root, getCommandRPC(yangmodules))
	installCompletionTree(cn0.tree.Root, getCommandOperState(yangmodules))
	installCompletionTree(cn0.tree.Root, getViewCommandConfig(yangmodules))
	installCompletionTree(cn1.tree.Root, getCommandConfig(yangmodules))
}

func installCompletionTree(root *CompletionNode, inputRoot *CompletionNode) error {
	excludeIndex := []int{}
	for _, child := range root.Childs {
		for idx, inputChild := range inputRoot.Childs {
			if child.Name == inputChild.Name {
				excludeIndex = append(excludeIndex, idx)
				if err := installCompletionTree(child, inputChild); err != nil {
					return err
				}
			}
		}
	}

	for idx, inputChild := range inputRoot.Childs {
		if !slicesContainsInt(excludeIndex, idx) {
			root.Childs = append(root.Childs, inputChild)
		}
	}
	return nil
}

type CompletionTree struct {
	Root *CompletionNode
}

type CompletionNode struct {
	Name        string
	Description string
	Childs      []*CompletionNode
}

func (n *CompletionNode) String() string {
	name := "<root>"
	if n.Name != "" {
		name = n.Name
	}
	return name
}

func (t *CompletionTree) Completion(line string, pos int) []*CompletionNode {
	line = line[:pos]
	args := strings.Fields(line)
	if len(args) == 0 {
		args = append(args, "")
	} else if line[len(line)-1] == ' ' {
		args = append(args, "")
	}
	log.Printf("DEBUG: line=\"%s\" pos=%d, args=%d:%+v", line, pos, len(args),
		args)

	search := func(nodes []*CompletionNode, str string) []*CompletionNode {
		result := []*CompletionNode{}
		for _, node := range nodes {
			switch node.Name {
			case "NAME":
				result = append(result, node)
			case "VALUE":
				result = append(result, node)

			// // TODO(slankdev)
			// case "INTEGER":
			// 	if _, err := strconv.Atoi(str); err == nil {
			// 		result = append(result, node)
			// 	} else if str == "" {
			// 		result = append(result, node)
			// 	}

			default:
				log.Printf("HasPrefix(%s,%s)\n", node.Name, str)
				if strings.HasPrefix(node.Name, str) {
					result = append(result, node)
				}
			}
		}
		return result
	}

	tree := getCommandNodeCurrent().tree
	var pivot *CompletionNode = tree.Root
	last := len(args) - 1
	for i, arg := range args {
		if i == last {
			log.Printf("loop (%d) last\n", i)
		} else {
			log.Printf("loop (%d)\n", i)
		}
		log.Printf("loop (%d) %s", i, pivot.String())

		candidates := search(pivot.Childs, arg)
		n := len(candidates)
		log.Printf("loop (%d) match %d candidates\n", i, n)

		if i == last {
			return candidates
		}

		switch {
		case n == 0:
			return nil
		case n == 1:
			pivot = candidates[0]
		case n >= 1:
			pivot = candidates[0]
		}
	}

	return nil
}

func DigNode(node *CompletionNode, query []string) *CompletionNode {
	if len(query) == 0 {
		return node
	}

	for idx, _ := range node.Childs {
		child := node.Childs[idx]
		if child.Name == query[0] {
			return DigNode(child, query[1:])
		}
	}
	return nil

}

func DigNodeOrDie(mode CliMode, query []string) *CompletionNode {
	root := DigNode(getCommandNode(mode).tree.Root, query)
	if root == nil {
		panic(fmt.Sprintf("Notfound %+v", query))
	}
	return root
}

func completer(line string, pos int) (string, []string, string) {
	tree := getCommandNodeCurrent().tree
	nodes := tree.Completion(line, pos)
	if len(nodes) == 0 {
		return line[:pos], nil, line[pos:]
	}

	names := []string{}
	for _, node := range nodes {
		names = append(names, node.Name)
	}

	pre := line[:pos]
	if len(pre) > 0 && pre[len(pre)-1] != ' ' {
		words := strings.Fields(pre)
		last := len(words) - 1
		words = append(words[:last], words[last+1:]...)
		pre = ""
		for _, word := range words {
			pre += word
			pre += " "
		}
	}

	if len(names) == 1 && names[0] == "<cr>" {
		return pre, nil, line[pos:]
	}

	if len(names) > 0 && names[len(names)-1] == "NAME" {
		return pre, nil, line[pos:]
	}
	if len(names) > 0 && names[len(names)-1] == "VALUE" {
		return pre, nil, line[pos:]
	}

	if len(names) == 1 {
		names[0] += " "
		return pre, names, line[pos:]
	}

	return pre, names, line[pos:]
}

func completionLister(line string, pos int) {
	tree := getCommandNodeCurrent().tree
	nodes := tree.Completion(line, pos)
	if len(nodes) == 0 {
		fmt.Fprintf(stdout, "\n%% Invalid input detected\n")
		return
	}

	longestnamelen := 0
	for _, node := range nodes {
		if len(node.Name) > longestnamelen {
			longestnamelen = len(node.Name)
		}
	}
	padding := func(str string, maxlen int) string {
		retStr := ""
		for i := 0; i < maxlen-len(str); i++ {
			retStr += " "
		}
		return retStr
	}

	fmt.Fprintf(stdout, "\nPossible Completions:\n")
	for _, node := range nodes {
		fmt.Fprintf(stdout, "  %s%s  %s\n", node.Name, padding(node.Name, longestnamelen),
			node.Description)
	}
}

type CommitHistory struct {
	Timestamp time.Time
	Before    string
	After     string
	Client    string
	Comment   string
}

func initCommitHistories() {
	if GlobalOptRunFilePath != "" {
		files, err := ioutil.ReadDir(GlobalOptRunFilePath)
		if err != nil {
			panic(err)
		}
		for _, file := range files {
			if strings.HasPrefix(file.Name(), "history") {
				fn := GlobalOptRunFilePath + "/" + file.Name()
				h, err := readCommitHistoryFromFile(fn)
				if err != nil {
					panic(err)
				}
				commitHistories = append([]CommitHistory{h}, commitHistories...)
			}
		}
	}
}

func readCommitHistoryFromFile(filename string) (CommitHistory, error) {
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
	if GlobalOptRunFilePath != "" {
		if err := h.WriteToFile(GlobalOptRunFilePath); err != nil {
			fmt.Fprintf(stdout, "Warning: %s ... ignored\n", err.Error())
		}
	}

	cliMode = CliModeConfigure
	dbm.root = *dbm.candidateRoot.DeepCopy()
	if err := dbm.root.WriteToJsonFile(getDatabasePath()); err != nil {
		fmt.Fprintf(stdout, "Error: %s\n", err.Error())
	}
}

func ccbShowConfigurationCommitList(args []string) {
	if len(args) > 4 {
		idx, err := strconv.Atoi(args[4])
		if err != nil {
			fmt.Fprintf(stdout, "Error: %s\n", err.Error())
			return
		}
		if len(commitHistories) < idx {
			fmt.Fprintf(stdout, "Error: Invalid commit idx\n")
			return
		}
		history := commitHistories[idx]
		na, err1 := ReadFromJsonString(history.Before)
		nb, err2 := ReadFromJsonString(history.After)
		if err1 != nil || err2 != nil {
			fmt.Fprintf(stdout, "Error1: %s\n", err1.Error())
			fmt.Fprintf(stdout, "Error2: %s\n", err2.Error())
			return
		}
		diff := DBNodeDiff(na, nb)
		fmt.Fprintln(stdout, diff)

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
		fmt.Fprintf(stdout, "Usage\n")
		return
	}

	idx, err := strconv.Atoi(args[4])
	if err != nil {
		fmt.Fprintf(stdout, "Error: %s\n", err.Error())
		return
	}
	if len(commitHistories) < idx {
		fmt.Fprintf(stdout, "Error: Invalid commit idx\n")
		return
	}

	history := commitHistories[idx]
	node, err := history.ToDBNode()
	if err != nil {
		fmt.Fprintf(stdout, "Error: %s\n", err.Error())
		return
	}

	fmt.Fprintln(stdout, DBNodeDiff(&dbm.root, node))
}

func ccbRollbackConfiguration(args []string) {
	if len(args) < 3 {
		fmt.Fprintf(stdout, "Usage: rollback configuration <idx>\n")
		return
	}
	idxArg := args[2]

	idx, err := strconv.Atoi(idxArg)
	if err != nil {
		fmt.Fprintf(stdout, "Error: %s\n", err.Error())
		return
	}
	if len(commitHistories) < idx {
		fmt.Fprintf(stdout, "Error: Invalid commit idx\n")
		return
	}

	history := commitHistories[idx]
	node, err := history.ToDBNode()
	if err != nil {
		fmt.Fprintf(stdout, "Error: %s\n", err.Error())
		return
	}

	dbm.candidateRoot = node.DeepCopy()
}

func dumpCompletionTreeJson(root *CompletionNode) string {
	out, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(out)
}

func yangModulesPath(path string) (*yang.Modules, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	modules := yang.NewModules()
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if !strings.HasSuffix(file.Name(), ".yang") {
			continue
		}
		fullname := fmt.Sprintf("%s/%s", path, file.Name())
		log.Printf("loading yang module '%s'\n", fullname)
		if err := modules.Read(fullname); err != nil {
			return nil, err
		}
	}

	errs := modules.Process()
	if len(errs) > 0 {
		for _, err := range errs {
			log.Printf("%s", err.Error())
		}
		return nil, errs[0]
	}
	return modules, nil
}
