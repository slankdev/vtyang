package vtyang

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/nsf/jsondiff"
	"github.com/openconfig/goyang/pkg/indent"
	"github.com/openconfig/goyang/pkg/yang"
	"github.com/pkg/errors"

	"github.com/slankdev/vtyang/pkg/util"
)

type DatabaseManager struct {
	modules *yang.Modules
	root    DBNode

	// candidateRoot is the top of candidate config
	candidateRoot *DBNode
}

func NewDatabaseManager() *DatabaseManager {
	m := DatabaseManager{}
	m.modules = yang.NewModules()
	m.candidateRoot = nil
	return &m
}

func (m *DatabaseManager) LoadDatabaseFromData(n *DBNode) error {
	m.root = *n
	return nil
}

func (m *DatabaseManager) LoadDatabaseFromFile(f string) error {
	if !util.FileExists(f) {
		if err := os.WriteFile(f, []byte("{}"), 0644); err != nil {
			return errors.Wrap(err, "writefile")
		}
	}

	root, err := ReadFromJsonFile(f)
	if err != nil {
		return errors.Wrap(err, "readfile")
	}
	m.root = *root
	return nil
}

func (m *DatabaseManager) LoadYangModuleOrDie(path string) {
	if err := m.LoadYangModule(path); err != nil {
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
		if !strings.HasSuffix(file.Name(), ".yang") {
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
		ent := yang.ToEntry(m)
		for _, e := range ent.Dir {
			entries = append(entries, e)
		}
	}
	return entries
}

func (dbm *DatabaseManager) GetNode(xpath XPath) (*DBNode, error) {
	n := &dbm.root
	xwords := xpath.words

	for ; len(xwords) != 0; xwords = xwords[1:] {
		xword := xwords[0]
		switch n.Type {
		case Container:
			found := false
			for idx := range n.Childs {
				child := &n.Childs[idx]
				if child.Name == xword.word {
					switch child.Type {
					case Leaf:
						fallthrough
					case Container:
						n = child
						found = true
						goto end
					case List:
						for idx2 := range child.Childs {
							child2 := &child.Childs[idx2]
							for idx3 := range child2.Childs {
								child3 := &child2.Childs[idx3]
								if xword.keys == nil {
									panic("database is broken")
								}
								for k, v := range xword.keys {
									if child3.Name == k && child3.Value.String == v {
										n = child2
										found = true
										goto end
									}
								}
							}
						}
					}
				}
			}
			if !found {
				return nil, nil
			}
		default:
			panic("UNSUPPORTED")
		}
	end:
	}

	return n, nil
}

func (dbm *DatabaseManager) DeleteNode(xpath XPath) error {
	n := dbm.candidateRoot
	xwords := xpath.words

	if len(xwords) == 0 {
		n.Childs = []DBNode{}
		return nil
	}

	for ; len(xwords) != 0; xwords = xwords[1:] {
		if n.Type != Container {
			panic(fmt.Sprintf("ASSERT(%s)", n.Type))
		}

		xword := xwords[0]
		found := false
		for idx := range n.Childs {
			child := &n.Childs[idx]
			if child.Name == xword.word {
				found = true
				switch child.Type {
				case Container:
					if len(xwords) == 1 {
						n.Childs = append(n.Childs[:idx], n.Childs[idx+1:]...)
						return nil
					}
					n = child
				case List:
					if xword.keys == nil {
						panic("database is broken")
					}
					cidx := lookupChildIdx(child, xword.keys)
					if cidx < 0 {
						return fmt.Errorf("not found (1)")
					}
					if len(xwords) == 1 {
						child.Childs = append(child.Childs[:cidx], child.Childs[cidx+1:]...)
						return nil
					}
					n = EnsureListNode(child, xword.keys)
				case Leaf:
					if len(xwords) != 1 {
						panic("ASSERT")
					}
					n.Childs = append(n.Childs[:idx], n.Childs[idx+1:]...)
					return nil
				default:
					panic(fmt.Sprintf("ASSERT(%s)", child.Type))
				}
			}
		}
		if !found {
			return fmt.Errorf("node not found (2)")
		}
	}
	return fmt.Errorf("node not found (3)")
}

func (dbm *DatabaseManager) SetNode(xpath XPath, val string) (
	*DBNode, error) {
	n := dbm.candidateRoot
	xwords := xpath.words
	for ; len(xwords) != 0; xwords = xwords[1:] {
		xword := xwords[0]
		switch xword.dbtype {
		case Container:
			found := false
			for idx := range n.Childs {
				child := &n.Childs[idx]
				if child.Name == xword.word {
					found = true
					switch child.Type {
					case Container:
						n = child
					case List:
						if xword.keys == nil {
							panic("database is broken")
						}
						listElement := EnsureListNode(child, xword.keys)
						if listElement == nil {
							panic("ASSERTION")
						}
						n = listElement
					case Leaf:
						if len(xwords) != 1 {
							panic("ASSERT")
						}
						if val != "" {
							(&(child.Value)).SetFromString(val)
						}
						return child, nil
					default:
						panic(fmt.Sprintf("ASSERT(%s)", child.Type))
					}
				}
				if found {
					break
				}
			}

			// not found case
			if !found {
				newnode := DBNode{Name: xword.word}
				newnode.Type = xword.dbtype
				if xword.keys != nil {
					for k, v := range xword.keys {
						newnode.Childs = []DBNode{
							{
								Type: Container,
								Childs: []DBNode{
									{
										Name: k,
										Type: Leaf,
										Value: DBValue{
											Type:   YString,
											String: v,
										},
									},
								},
							},
						}
					}
				}
				if xword.dbtype == Leaf {
					newnode.Value = DBValue{}
					newnode.Value.Type = xword.dbvaluetype
					(&newnode.Value).SetFromString(val)
				}

				n.Childs = append(n.Childs, newnode)
				n = &n.Childs[len(n.Childs)-1]
			}
		case List:
			// Ensure List's leaf
			found1 := false
			for idx := range n.Childs {
				if n.Childs[idx].Name == xword.word {
					found1 = true
					n = &n.Childs[idx]
					break
				}
			}
			if !found1 {
				n.Childs = append(n.Childs, DBNode{
					Name: xword.word,
					Type: List,
				})
				n = &n.Childs[len(n.Childs)-1]
			}

			// Ensure List-Key leaf
			found2 := false
			for idx := range n.Childs {
				match := true
				for k, v := range xword.keys {
					for _, c := range n.Childs[idx].Childs {
						if c.Name == k && c.Value.String != v {
							match = false
						}
					}
				}
				if match {
					found2 = true
					n = &n.Childs[idx]
					break
				}
			}
			if !found2 {
				listChilds := []DBNode{}
				for k, v := range xword.keys {
					listChilds = append(listChilds, DBNode{
						Name: k,
						Type: Leaf,
						Value: DBValue{
							Type:   YString,
							String: v,
						},
					})
				}
				n.Childs = append(n.Childs, DBNode{
					Type:   Container,
					Childs: listChilds,
				})
				n = &n.Childs[len(n.Childs)-1]
			}
		case Leaf:
			switch xword.dbvaluetype {
			case YInteger:
				ival, err := strconv.Atoi(val)
				if err != nil {
					return nil, err
				}
				n.Childs = append(n.Childs, DBNode{
					Name: xword.word,
					Type: Leaf,
					Value: DBValue{
						Type:    YInteger,
						Integer: ival,
					},
				})
			case YString:
				n.Childs = append(n.Childs, DBNode{
					Name: xword.word,
					Type: Leaf,
					Value: DBValue{
						Type:   YString,
						String: val,
					},
				})
			default:
				return nil, fmt.Errorf("%s: unsupported(%s)", util.LINE(), xword.dbvaluetype)
			}
		case LeafList:
			var tmpNode *DBNode
			for idx := range n.Childs {
				if n.Childs[idx].Name == xword.word {
					tmpNode = &n.Childs[idx]
					break
				}
			}
			if tmpNode == nil {
				n.Childs = append(n.Childs, DBNode{
					Name: xword.word,
					Type: LeafList,
				})
				tmpNode = &n.Childs[len(n.Childs)-1]
			}
			tmpNode.Value = DBValue{
				Type:        YStringArray,
				StringArray: strings.Fields(val),
			}
		default:
			return nil, fmt.Errorf("%s: unsupported(%s)", util.LINE(), xword.dbtype)
		}
	}

	return n, nil
}

func (xpath XPath) CreateDBNodeTree() (*DBNode, error) {
	root := DBNode{
		Name: "<root>",
		Type: Container,
	}

	var tail *DBNode = &root
	for _, xword := range xpath.words {
		n := new(DBNode)
		n.Name = xword.word
		n.Type = Container

		if xword.keys != nil {
			n.Type = List
			n.Childs = []DBNode{
				{Type: Container},
			}
			for k, v := range xword.keys {
				n.Childs[0].Childs = []DBNode{
					{
						Name: k,
						Type: Leaf,
						Value: DBValue{
							Type:   YString,
							String: v,
						},
					},
				}
			}
		}

		tail.Childs = append(tail.Childs, *n)
		tail = &tail.Childs[len(tail.Childs)-1]
	}

	//root.Write(os.Stdout)
	return &root, nil
}

func EnsureListNode(listNode *DBNode, kv map[string]string) *DBNode {
	if listNode.Type != List {
		panic("ASSERTION")
	}

	for idx := range listNode.Childs {
		elementRoot := &listNode.Childs[idx]
		if matchChild(elementRoot, kv) {
			return elementRoot
		}
	}

	newElement := DBNode{Type: Container}
	for k, v := range kv {
		n := DBNode{
			Name: k,
			Type: Leaf,
			Value: DBValue{
				Type:   YString,
				String: v,
			},
		}
		newElement.Childs = append(newElement.Childs, n)
	}

	listNode.Childs = append(listNode.Childs, newElement)
	return &listNode.Childs[len(listNode.Childs)-1]
}

func matchChild(root *DBNode, kv map[string]string) bool {
	nMatch := 0
	for idx := range root.Childs {
		child := &root.Childs[idx]
		for k, v := range kv {
			if child.Name == k && child.Value.String == v {
				nMatch++
			}
		}
	}
	return len(kv) == nMatch
}

func lookupChildIdx(root *DBNode, kv map[string]string) int {
	for idx := range root.Childs {
		child := &root.Childs[idx]
		for idx2 := range child.Childs {
			child2 := &child.Childs[idx2]
			nMatch := 0
			for k, v := range kv {
				if child2.Name == k && child2.Value.String == v {
					nMatch++
				}
			}
			if len(kv) == nMatch {
				return idx
			}
		}
	}
	return -1
}

func (m *DatabaseManager) Dump() {
	entries := m.DumpEntries()
	for _, e := range entries {
		dump(os.Stdout, e)
	}
}

func getTypeName(e *yang.Entry) string {
	if e == nil || e.Type == nil {
		return ""
	}
	return e.Type.Root.Name
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

type DBNodeType string

const (
	Container DBNodeType = "container"
	List      DBNodeType = "list"
	Leaf      DBNodeType = "leaf"
	LeafList  DBNodeType = "leaf-list"
)

type DBNode struct {
	Name   string
	Type   DBNodeType
	Childs []DBNode
	Value  DBValue
}

type DBValueType string

const (
	YString      DBValueType = "string"
	YInteger     DBValueType = "integer"
	YBoolean     DBValueType = "boolean"
	YStringArray DBValueType = "stringarray"
)

type DBValue struct {
	Type DBValueType

	// Union
	Integer     int
	String      string
	Boolean     bool
	StringArray []string
}

func (n *DBNode) DeepCopy() *DBNode {
	m := n.ToMap()
	copy, _ := Interface2DBNode(m)
	return copy
}

func (n *DBNode) String() string {
	if m := n.ToMap(); m == nil {
		return "{}"
	}
	return js(n.ToMap())
}

func (n *DBNode) ToMap() interface{} {
	m := map[string]interface{}{}
	switch n.Type {
	case Container:
		for _, child := range n.Childs {
			m[child.Name] = child.ToMap()
		}
	case List:
		array := []interface{}{}
		for _, child := range n.Childs {
			array = append(array, child.ToMap())
		}
		return array
	case Leaf:
		return n.Value.ToValue()
	case LeafList:
		return n.Value.ToValue()
	case "":
		return nil
	default:
		panic(fmt.Sprintf("ASSERT(%s)", n.Type))
	}
	return m
}

func (n *DBNode) WriteToJsonFile(filename string) error {
	s := dbm.root.String()
	if err := ioutil.WriteFile(filename, []byte(s), 0644); err != nil {
		return err
	}
	return nil
}

func ReadFromJsonString(jsonstr string) (*DBNode, error) {
	m := map[string]interface{}{}
	if err := json.Unmarshal([]byte(jsonstr), &m); err != nil {
		return nil, err
	}
	return Interface2DBNode(m)
}

func ReadFromJsonFile(filename string) (*DBNode, error) {
	raw, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return ReadFromJsonString(string(raw))
}

func Interface2DBNode(i interface{}) (*DBNode, error) {
	n := &DBNode{}
	switch g := i.(type) {
	case map[string]interface{}:
		for k, v := range g {
			child, err := Interface2DBNode(v)
			if err != nil {
				return nil, err
			}
			n.Type = Container
			child.Name = k
			n.Childs = append(n.Childs, *child)
		}
	case []interface{}:
		isLeafList := false
		childCandidates := []DBNode{}
		for _, v := range g {
			child, err := Interface2DBNode(v)
			if err != nil {
				return nil, err
			}
			if child.Type == Leaf {
				isLeafList = true
			}
			childCandidates = append(childCandidates, *child)
		}
		if isLeafList {
			v, err := squashListToLeafList(childCandidates)
			if err != nil {
				return nil, err
			}
			n.Type = LeafList
			n.Value = v
		} else {
			n.Type = List
			n.Childs = append(n.Childs, childCandidates...)
		}
	case bool:
		n.Type = Leaf
		n.Value = DBValue{
			Type:    YBoolean,
			Boolean: g,
		}
	case int:
		n.Type = Leaf
		n.Value = DBValue{
			Type:    YInteger,
			Integer: g,
		}
	case float64:
		n.Type = Leaf
		n.Value = DBValue{
			Type:    YInteger,
			Integer: int(g),
		}
	case string:
		n.Type = Leaf
		n.Value = DBValue{
			Type:   YString,
			String: g,
		}
	case []string:
		n.Type = LeafList
		n.Value = DBValue{
			Type:        YStringArray,
			StringArray: g,
		}
	case nil:
		n.Type = Container
	default:
		panic(fmt.Sprintf("ASSERT(%T)", g))
	}
	return n, nil
}

func (v DBValue) ToValue() interface{} {
	switch v.Type {
	case YInteger:
		return v.Integer
	case YBoolean:
		return v.Boolean
	case YString:
		return v.String
	case YStringArray:
		return v.StringArray
	default:
		panic(fmt.Sprintf("ASSERT(%s)", v.Type))
	}
}

func (v *DBValue) SetFromString(s string) error {
	switch v.Type {
	case YInteger:
		i, err := strconv.Atoi(s)
		if err != nil {
			return err
		}
		v.Integer = i
	case YBoolean:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		v.Boolean = b
	case YString:
		v.String = s
	}
	return nil
}

func DBNodeDiff(na, nb *DBNode) string {
	a := []byte(na.String())
	b := []byte(nb.String())
	opts := jsondiff.DefaultConsoleOptions()
	opts.Indent = "  "
	opt, diff := jsondiff.Compare(a, b, &opts)
	if opt == jsondiff.FullMatch {
		return ""
	}
	return diff
}

func filterDbWithModule(n *DBNode, modName string) (*DBNode, error) {
	var mod *yang.Module = nil
	for fullname, m := range yangmodules.Modules {
		if fullname == modName {
			mod = m
		}
	}
	if mod == nil {
		return nil, fmt.Errorf("module(%s) not found", modName)
	}
	ret := n.DeepCopy()
	filteredChild := []DBNode{}
	for _, e := range yang.ToEntry(mod).Dir {
		if e.ReadOnly() ||
			e.RPC != nil ||
			e.Kind == yang.NotificationEntry {
			continue
		}
		for _, child := range ret.Childs {
			if child.Name == e.Name {
				filteredChild = append(filteredChild, *filterDbWithModuleImpl(&child, e))
			}
		}
	}
	ret.Childs = filteredChild

	// Append modName
	for idx := range ret.Childs {
		ret.Childs[idx].Name = fmt.Sprintf("%s:%s", modName, ret.Childs[idx].Name)

	}
	return ret, nil
}

func filterDbWithModuleImpl(n *DBNode, root *yang.Entry) *DBNode {
	childs := []DBNode{}
	switch {
	case root.IsList():
		//fmt.Println("LIST")
		for _, nn := range n.Childs {
			childs2 := []DBNode{}
			for _, e := range root.Dir {
				for _, child := range nn.Childs {
					if child.Name == e.Name {
						childs2 = append(childs2, *filterDbWithModuleImpl(&child, e))
					}
				}
			}
			childs = append(childs, DBNode{
				Name:   "",
				Type:   Container,
				Childs: childs2,
			})
		}
	case root.IsContainer():
		//fmt.Println("CONTAINER")
		for _, e := range root.Dir {
			for _, child := range n.Childs {
				if child.Name == e.Name {
					childs = append(childs, *filterDbWithModuleImpl(&child, e))
				}
			}
		}
	case root.IsLeafList():
		//fmt.Println("LeafList", root.Name, n.Name)
		if root.Name == n.Name {
			childs = append(childs, *n)
		}
	case root.IsLeaf():
		//fmt.Println("LEAF")
		if root.Name == n.Name {
			childs = append(childs, *n)
		}
	default:
		panic(fmt.Sprintf("OKASHII %s", root.Kind))
	}
	n.Childs = childs
	return n
}

func squashListToLeafList(items []DBNode) (DBValue, error) {
	typesMap := map[DBValueType]bool{}
	for _, item := range items {
		typesMap[item.Value.Type] = true
	}
	types := []DBValueType{}
	for key, val := range typesMap {
		if val {
			types = append(types, key)
		}
	}
	if len(types) != 1 {
		return DBValue{}, fmt.Errorf("invalid list items (%+v)", types)
	}

	switch types[0] {
	// TODO(slankdev): implemente me
	// case YInteger:
	// case YBoolean:

	case YString:
		g := []string{}
		for _, item := range items {
			g = append(g, item.Value.String)
		}
		ret := DBValue{
			Type:        YStringArray,
			StringArray: g,
		}
		return ret, nil
	default:
		panic(fmt.Sprintf("OKASHII %s", types[0]))
	}
}
