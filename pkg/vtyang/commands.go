package vtyang

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/nsf/jsondiff"
	"github.com/openconfig/goyang/pkg/yang"
	"github.com/pkg/errors"

	"github.com/slankdev/vtyang/pkg/mgmtd"
	"github.com/slankdev/vtyang/pkg/util"
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

			if agentOpts.BackendMgmtd != nil {
				// Un-Lock mgmtd datastores
				for _, dsId := range []*mgmtd.DatastoreId{
					mgmtd.DatastoreId_RUNNING_DS.Enum(),
					mgmtd.DatastoreId_CANDIDATE_DS.Enum(),
				} {
					if err := mgmtdClient.LockReq(&mgmtd.FeLockDsReq{
						SessionId: mgmtdClient.GetSessionId(),
						ReqId:     util.NewUint64Pointer(0),
						DsId:      dsId,
						Lock:      util.NewBoolPointer(false),
					}); err != nil {
						err := errors.Wrap(err, "LockReq(Lock=false)")
						fmt.Fprintf(stdout, "Error %v\n", err)
						return
					}
				}
			}
		}
	})

	installCommand(mode, "show cli-tree", []string{
		"Display information",
		"Display completion tree",
	}, func(arg []string) {
		fmt.Fprintln(stdout, dumpCompletionTreeJson(getCommandNodeCurrent().tree.Root))
	})

	installCommandNoCompletion(mode, "hidden-command-nothing", func(arg []string) {
		// do nothing
	})

	installCommand(mode, "save cli-tree", []string{
		"Save information",
		"Save completion tree",
	}, func(arg []string) {
		content := dumpCompletionTreeJson(getCommandNodeCurrent().tree.Root)
		if err := os.WriteFile("/tmp/clitree.json", []byte(content), os.ModePerm); err != nil {
			fmt.Fprintf(stdout, "ERROR: %s", err.Error())
		}
	})

	installCommandNoCompletion(mode, "show-xpath", func(args []string) {
		xpath, _, err := ParseXPathArgs(dbm, args[1:], true)
		if err != nil {
			fmt.Fprintf(stdout, "Error: %s\n", err.Error())
			return
		}
		out, err := json.MarshalIndent(xpath, "", "  ")
		if err != nil {
			fmt.Fprintf(stdout, "Error: %s\n", err.Error())
			return
		}
		fmt.Fprintf(stdout, "%s\n", string(out))
	})

	installCommandNoCompletion(mode, "eval-cli", func(args []string) {
		xpath, val, tail, err := ParseXPathCli(dbm, args[1:], []string{}, true)
		if err != nil {
			fmt.Fprintf(stdout, "Error: %s\n", err.Error())
			return
		}
		out, err := json.MarshalIndent(struct {
			XPath XPath
			Value []DBValue `json:",omitempty"`
			Tail  []string  `json:",omitempty"`
		}{
			XPath: xpath,
			Value: val,
			Tail:  tail,
		}, "", "  ")
		if err != nil {
			fmt.Fprintf(stdout, "Error: %s\n", err.Error())
			return
		}
		fmt.Fprintf(stdout, "%s\n", string(out))
	})

	installCommandNoCompletion(mode, "eval-xpath", func(args []string) {
		if len(args) != 2 {
			err := errors.Errorf("Usage: %s <xpath>", args[0])
			fmt.Fprintf(stdout, "Error: %s\n", err)
			return
		}
		xp, err := ParseXPathString(dbm, args[1])
		if err != nil {
			fmt.Fprintf(stdout, "Error: %s\n", err)
			return
		}
		out, err := json.MarshalIndent(xp, "", "  ")
		if err != nil {
			fmt.Fprintf(stdout, "Error: %s\n", err.Error())
			return
		}
		fmt.Fprintf(stdout, "%s\n", string(out))
	})
}

func installCommandNoCompletion(mode CliMode, match string,
	f func(args []string)) {
	cn := getCommandNode(mode)
	cn.commands = append(cn.commands, Command{m: match, f: f})
}

func installCommand(mode CliMode, match string, helps []string,
	f func(args []string)) {
	installCommandNoCompletion(mode, match, f)

	cn := getCommandNode(mode)
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

			if agentOpts.BackendMgmtd != nil {
				// Lock mgmtd datastores
				for _, dsId := range []*mgmtd.DatastoreId{
					mgmtd.DatastoreId_RUNNING_DS.Enum(),
					mgmtd.DatastoreId_CANDIDATE_DS.Enum(),
				} {
					if err := mgmtdClient.LockReq(&mgmtd.FeLockDsReq{
						SessionId: mgmtdClient.GetSessionId(),
						ReqId:     util.NewUint64Pointer(0),
						DsId:      dsId,
						Lock:      util.NewBoolPointer(true),
					}); err != nil {
						err := errors.Wrap(err, "LockReq(Lock=true)")
						fmt.Fprintf(stdout, "Error %v\n", err)
						return
					}
				}
			}
		})

	installCommand(CliModeConfigure,
		"show configuration running", []string{
			"Display information",
			"Display configuration information",
			"Display running-configuration information",
		},
		func(args []string) {
			dsId := mgmtd.DatastoreId_RUNNING_DS.Enum()
			config, err := mgmtdClient.GetReq(&mgmtd.FeGetReq{
				SessionId: mgmtdClient.GetSessionId(),
				Config:    util.NewBoolPointer(true),
				DsId:      dsId,
				ReqId:     util.NewUint64Pointer(0),
				Data: []*mgmtd.YangGetDataReq{
					{
						Data: &mgmtd.YangData{
							Xpath: util.NewStringPointer("/"),
						},
						NextIndx: util.NewInt64Pointer(0),
					},
				},
			})
			if err != nil {
				err := errors.Wrap(err, "mgmtd.GetReq")
				fmt.Fprintf(stdout, "Error: %v", err)
				return
			}
			configJson, err := frrConfigToJson(config)
			if err != nil {
				err := errors.Wrap(err, "frrConfigToJson")
				fmt.Fprintf(stdout, "Error: %s\n", err)
				return
			}
			fmt.Fprintf(stdout, "%s\n", configJson)
		})

	installCommand(CliModeConfigure,
		"show configuration candidate", []string{
			"Display information",
			"Display configuration information",
			"Display candidate-configuration information",
		},
		func(args []string) {
			dsId := mgmtd.DatastoreId_CANDIDATE_DS.Enum()
			config, err := mgmtdClient.GetReq(&mgmtd.FeGetReq{
				SessionId: mgmtdClient.GetSessionId(),
				Config:    util.NewBoolPointer(true),
				DsId:      dsId,
				ReqId:     util.NewUint64Pointer(0),
				Data: []*mgmtd.YangGetDataReq{
					{
						Data: &mgmtd.YangData{
							Xpath: util.NewStringPointer("/"),
						},
						NextIndx: util.NewInt64Pointer(0),
					},
				},
			})
			if err != nil {
				err := errors.Wrap(err, "mgmtd.GetReq")
				fmt.Fprintf(stdout, "Error: %s\n", err)
				return
			}
			configJson, err := frrConfigToJson(config)
			if err != nil {
				err := errors.Wrap(err, "frrConfigToJson")
				fmt.Fprintf(stdout, "Error: %s\n", err)
				return
			}
			fmt.Fprintf(stdout, "%s\n", configJson)
		})

	installCommand(CliModeConfigure,
		"show configuration diff", []string{
			"Display information",
			"Display configuration information",
			"Display configuration diff",
		},
		func(args []string) {
			if agentOpts.BackendMgmtd == nil {
				diff := DBNodeDiff(&dbm.root, dbm.candidateRoot)
				fmt.Fprintln(stdout, diff)
			} else {
				diff, err := frrConfigDiff()
				if err != nil {
					fmt.Fprintf(stdout, "Error: %s\n", err)
				}
				fmt.Fprintln(stdout, diff)
			}
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
	Modules     []string
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
	names := []string{}
	result := doCompletion(line, pos)
	for _, item := range result.Items {
		names = append(names, item.Word)
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

type CompletionResult struct {
	// InvalidArg
	InvalidArg bool
	// ResolvedXPath
	ResolvedXPath *XPath
	// Items
	Items []CompletionItem `json:",omitempty"`
}

type CompletionItem struct {
	Word   string `json:",omitempty"`
	Helper string `json:",omitempty"`
}

func doCompletion(line string, pos int) CompletionResult {
	ret := CompletionResult{}
	nodes := getCommandNodeCurrent().tree.Completion(line, pos)
	items := []CompletionItem{}

	tailSpace := false
	if len(line) > 0 {
		tailSpace = line[pos-1] == ' '
	}

	// Static information (1) yang-tree
	for _, node := range nodes {
		items = append(items, CompletionItem{
			Word:   node.Name,
			Helper: node.Description,
		})
	}

	// Static information (2) identityref, enum
	args := strings.Fields(line)
	if len(args) > 1 {
		xpath, value, tail0, err := ParseXPathCli(dbm, args[1:], []string{}, true)
		if err != nil {
			fmt.Fprintf(stdout, "Error(%s): %s \n", util.LINE(), err)
			return CompletionResult{}
		}

		// (2.1) TailIsList
		eliminateName := false
		if xpath.TailIsList() {
			t := xpath.Tail()
			tailArg := args[len(args)-1]
			if tailSpace {
				eliminateName = true
				for _, k := range t.KeysIndex {
					if t.Keys[k].Value.Type == yang.Ynone {
						items0, err := resolveListKeyCompletionItems("", t.Keys[k].ytype)
						if err != nil {
							fmt.Fprintf(stdout, "Error(%s): %s \n", util.LINE(), err)
							return CompletionResult{}
						}
						items = append(items, items0...)
						break
					}
				}
			} else {
				eliminateName = true
				for _, k := range t.KeysIndex {
					if t.Keys[k].Value.Type == yang.Ynone {
						items0, err := resolveListKeyCompletionItems(tailArg, t.Keys[k].ytype)
						if err != nil {
							fmt.Fprintf(stdout, "Error(%s): %s \n", util.LINE(), err)
							return CompletionResult{}
						}
						items = append(items, items0...)
						break
					} else {
						value := t.Keys[k].Value
						if tailArg != value.ToString() {
							continue
						}
						items0, err := resolveListKeyCompletionItems(tailArg, t.Keys[k].ytype)
						if err != nil {
							fmt.Fprintf(stdout, "Error(%s): %s \n", util.LINE(), err)
							return CompletionResult{}
						}
						items = append(items, items0...)
						break
					}
				}
			}
		}

		// (2.2) TailIsLeaf
		eliminateValue := false
		if xpath.TailIsLeaf() {
			if tailSpace {
				if len(value) == 0 {
					ret.ResolvedXPath = &xpath
					if ret.ResolvedXPath != nil {
						if len(xpath.Words) > 0 {
							tail := xpath.Words[len(xpath.Words)-1]
							tailArg := ""
							if len(tail0) > 0 {
								tailArg = tail0[0]
							}
							items0, err := resolveLeafValueCompletionItems(tailArg, tail.ytype)
							if err != nil {
								fmt.Fprintf(stdout, "Error(%s): %s \n", util.LINE(), err)
								return CompletionResult{}
							}
							items = append(items, items0...)
						}
					}
				}
				eliminateValue = true
			} else {
				if value != nil {
					tail := xpath.Words[len(xpath.Words)-1]
					tailArg := args[len(args)-1]
					items0, err := resolveLeafValueCompletionItems(tailArg, tail.ytype)
					if err != nil {
						fmt.Fprintf(stdout, "Error(%s): %s \n", util.LINE(), err)
						return CompletionResult{}
					}
					items = append(items, items0...)
					eliminateValue = true
				} else {
					if len(tail0) > 0 {
						tail := xpath.Words[len(xpath.Words)-1]
						tailArg := tail0[0]
						items0, err := resolveLeafValueCompletionItems(tailArg, tail.ytype)
						if err != nil {
							fmt.Fprintf(stdout, "Error(%s): %s \n", util.LINE(), err)
							return CompletionResult{}
						}
						items = append(items, items0...)
						eliminateValue = true
					}
				}
			}
		}

		if eliminateName {
			for idx := range items {
				if items[idx].Word == "NAME" {
					items = append(items[:idx], items[idx+1:]...)
					break
				}
			}
		}
		if eliminateValue {
			for idx := range items {
				if items[idx].Word == "VALUE" {
					items = append(items[:idx], items[idx+1:]...)
					break
				}
			}
		}
	}

	// Unique
	uniqueItems := []CompletionItem{}
	for _, item := range items {
		found := false
		for _, t := range uniqueItems {
			if t.Word == item.Word {
				found = true
				break
			}
		}
		if !found {
			uniqueItems = append(uniqueItems, item)
		}
	}
	items = uniqueItems

	// Dynamic information
	// TODO(slankdev): implement me
	//
	// [TO-BE]
	// > set lib interface dum0 description hoge-dum0
	// > set lib interface dum4 description hoge-dum4
	// > set values items item4 ?
	// Pissible Completions
	//   NAME   interface name
	//   dum0   configured in candidate-ds
	//   dum1   configured in running-ds
	//   dum4   configured in candidate-ds

	// Return result
	sort.Slice(items, func(i, j int) bool {
		return items[i].Word < items[j].Word
	})
	ret.Items = items
	return ret
}

func completionLister(line string, pos int) {
	result := doCompletion(line, pos)
	if len(result.Items) == 0 {
		fmt.Fprintf(stdout, "\n%% Invalid input detected\n")
		return
	}

	longestnamelen := 0
	for _, item := range result.Items {
		if len(item.Word) > longestnamelen {
			longestnamelen = len(item.Word)
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
	for _, item := range result.Items {
		fmt.Fprintf(stdout, "  %s%s  %s\n", item.Word,
			padding(item.Word, longestnamelen),
			item.Helper)
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
		files, err := os.ReadDir(GlobalOptRunFilePath)
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
	b, err := os.ReadFile(filename)
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
	if err := os.WriteFile(filename, b, 0644); err != nil {
		return err
	}
	return nil
}

func ccbCommitCallback(args []string) {
	if agentOpts.BackendMgmtd != nil {
		if err := mgmtdClient.CommitConfig(&mgmtd.FeCommitConfigReq{
			SessionId:    mgmtdClient.GetSessionId(),
			ReqId:        util.NewUint64Pointer(0),
			SrcDsId:      mgmtd.DatastoreId_CANDIDATE_DS.Enum(),
			DstDsId:      mgmtd.DatastoreId_RUNNING_DS.Enum(),
			ValidateOnly: util.NewBoolPointer(false),
			Abort:        util.NewBoolPointer(false),
		}); err != nil {
			err := errors.Wrap(err, "mgmtd.WriteProtoBufMsg")
			fmt.Fprintf(stdout, "Error: %s", err)
		}
	}

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

func yangModulesPath(paths []string) (*yang.Modules, error) {
	modules := yang.NewModules()
	for _, path := range paths {
		files, err := os.ReadDir(path)
		if err != nil {
			return nil, err
		}
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

func frrConfigToJson(config []*mgmtd.YangData) ([]byte, error) {
	yd := []YangData{}
	for _, data := range config {
		xp, err := ParseXPathString(dbm, *data.Xpath)
		if err != nil {
			return nil, errors.Wrap(err, "ParseXPathString")
		}
		yd = append(yd, YangData{
			XPath: xp,
			Value: data.Value.GetEncodedStrVal(),
		})
	}

	out, err := CraftDBNode(yd)
	if err != nil {
		return nil, errors.Wrap(err, "CraftDBNode")
	}
	return []byte(out.String()), nil
}

func frrConfigDiff() (string, error) {
	configs := [2][]*mgmtd.YangData{}
	for idx, dsId := range []*mgmtd.DatastoreId{
		mgmtd.DatastoreId_RUNNING_DS.Enum(),
		mgmtd.DatastoreId_CANDIDATE_DS.Enum(),
	} {
		_, _ = idx, dsId
		config, err := mgmtdClient.GetReq(&mgmtd.FeGetReq{
			SessionId: mgmtdClient.GetSessionId(),
			Config:    util.NewBoolPointer(true),
			DsId:      dsId,
			ReqId:     util.NewUint64Pointer(0),
			Data: []*mgmtd.YangGetDataReq{
				{
					Data: &mgmtd.YangData{
						Xpath: util.NewStringPointer("/"),
					},
					NextIndx: util.NewInt64Pointer(0),
				},
			},
		})
		if err != nil {
			err := errors.Wrap(err, "mgmtd.GetReq")
			return "", err
		}
		configs[idx] = config
	}

	jsonRunningConfig, err := frrConfigToJson(configs[0])
	if err != nil {
		return "", err
	}
	jsonCandidateConfig, err := frrConfigToJson(configs[1])
	if err != nil {
		return "", err
	}
	opts := jsondiff.DefaultConsoleOptions()
	opts.Indent = "  "
	_, diff := jsondiff.Compare(jsonRunningConfig,
		jsonCandidateConfig, &opts)
	return diff, nil
}

func resolveListKeyCompletionItems(tailArg string,
	ytype yang.YangType) ([]CompletionItem, error) {
	return resolveCompletionItemsImpl(tailArg, ytype, false)
}

func resolveLeafValueCompletionItems(tailArg string,
	ytype yang.YangType) ([]CompletionItem, error) {
	return resolveCompletionItemsImpl(tailArg, ytype, true)
}

func resolveCompletionItemsImpl(tailArg string,
	ytype yang.YangType, isValue bool) ([]CompletionItem, error) {
	word := "NAME"
	if isValue {
		word = "VALUE"
	}

	// COMPLETION SWITCHES
	ok := false
	items := []CompletionItem{}

	// TYPE: identityref
	if ytype.IdentityBase != nil {
		for _, vv := range ytype.IdentityBase.Values {
			m, err := yang.ToEntry(vv.Parent).InstantiatingModule()
			if err != nil {
				return nil, errors.Wrap(err, "yang.ToEntry")
			}
			tmpName := fmt.Sprintf("%s:%s", m, vv.Name)
			if strings.HasPrefix(tmpName, tailArg) {
				items = append(items, CompletionItem{
					Word:   tmpName,
					Helper: "",
				})
			}
		}
		ok = true
	}

	// TYPE: string
	if ytype.Kind == yang.Ystring {
		items = append(items, CompletionItem{
			Word:   word,
			Helper: "",
		})
		ok = true
	}

	// TYPE: leafref
	if ytype.Kind == yang.Yleafref {
		items = append(items, CompletionItem{
			Word:   word,
			Helper: "TODO(leafref)",
		})
		ok = true
	}

	// TYPE: uint64
	if ytype.Kind == yang.Yuint64 {
		items = append(items, CompletionItem{
			Word:   word,
			Helper: "TODO(uint64)",
		})
		ok = true
	}

	// TYPE: uint32
	if ytype.Kind == yang.Yuint32 {
		items = append(items, CompletionItem{
			Word:   word,
			Helper: "TODO(uint32)",
		})
		ok = true
	}

	// TYPE: uint16
	if ytype.Kind == yang.Yuint16 {
		items = append(items, CompletionItem{
			Word:   word,
			Helper: "TODO(uint16)",
		})
		ok = true
	}

	// TYPE: uint8
	if ytype.Kind == yang.Yuint8 {
		items = append(items, CompletionItem{
			Word:   word,
			Helper: "TODO(uint8)",
		})
		ok = true
	}

	// TYPE: union
	if ytype.Kind == yang.Yunion {
		tmp := []CompletionItem{}
		for _, subytype := range ytype.Type {
			items0, err := resolveCompletionItemsImpl(tailArg, *subytype, isValue)
			if err != nil {
				return nil, errors.Wrap(err, "resolveCompletionItemsImpl(sub)")
			}
			tmp = append(tmp, items0...)
		}

		// Make unique for tmp
		ret := []CompletionItem{}
		for _, item := range tmp {
			found := false
			for _, t := range ret {
				if t.Word == item.Word {
					found = true
					break
				}
			}
			if !found {
				ret = append(ret, item)
			}
		}
		items = append(items, ret...)
		ok = true
	}

	// TYPE: enumeration
	if ytype.Kind == yang.Yenum {
		for _, n := range ytype.Enum.Names() {
			if strings.HasPrefix(n, tailArg) {
				items = append(items, CompletionItem{
					Word:   n,
					Helper: "",
				})
			}
		}
		ok = true
	}

	// ERROR CASE
	if !ok {
		fmt.Printf("\nSOMETHING UN-CATCHED 1 (%s)\n", ytype.Kind.String())
	}

	return items, nil
}
