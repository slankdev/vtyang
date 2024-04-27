package vtyang

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"reflect"
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
	files, err := os.ReadDir(path)
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
	xwords := xpath.Words

	for ; len(xwords) != 0; xwords = xwords[1:] {
		xword := xwords[0]
		switch n.Type {
		case Container:
			found := false
			for idx := range n.Childs {
				child := &n.Childs[idx]
				if child.Name == xword.Word {
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
								if xword.Keys == nil {
									panic("database is broken")
								}
								for k, v := range xword.Keys {
									if child3.Name == k && reflect.DeepEqual(child3.Value, v) {
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
	xwords := xpath.Words

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
			if child.Name == xword.Word {
				found = true
				switch child.Type {
				case Container:
					if len(xwords) == 1 {
						n.Childs = append(n.Childs[:idx], n.Childs[idx+1:]...)
						return nil
					}
					n = child
				case List:
					if xword.Keys == nil {
						panic("database is broken")
					}
					cidx := lookupChildIdx(child, xword.Keys)
					if cidx < 0 {
						return fmt.Errorf("not found (1)")
					}
					if len(xwords) == 1 {
						child.Childs = append(child.Childs[:cidx], child.Childs[cidx+1:]...)
						return nil
					}
					n = EnsureListNode(child, xword.Keys)
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
	xwords := xpath.Words
	for ; len(xwords) != 0; xwords = xwords[1:] {
		xword := xwords[0]
		switch xword.Dbtype {
		case Container:
			found := false
			for idx := range n.Childs {
				child := &n.Childs[idx]
				if child.Name == xword.Word {
					found = true
					switch child.Type {
					case Container:
						n = child
					case List:
						if xword.Keys == nil {
							panic("database is broken")
						}
						listElement := EnsureListNode(child, xword.Keys)
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
				newnode := DBNode{Name: xword.Word}
				newnode.Type = xword.Dbtype
				if xword.Keys != nil {
					for k, v := range xword.Keys {
						newnode.Childs = []DBNode{
							{
								Type: Container,
								Childs: []DBNode{
									{
										Name:  k,
										Type:  Leaf,
										Value: v,
									},
								},
							},
						}
					}
				}
				if xword.Dbtype == Leaf {
					newnode.Value = DBValue{}
					newnode.Value.Type = xword.Dbvaluetype
					(&newnode.Value).SetFromString(val)
				}

				n.Childs = append(n.Childs, newnode)
				n = &n.Childs[len(n.Childs)-1]
			}
		case List:
			// Ensure List's leaf
			found1 := false
			for idx := range n.Childs {
				if n.Childs[idx].Name == xword.Word {
					found1 = true
					n = &n.Childs[idx]
					break
				}
			}
			if !found1 {
				n.Childs = append(n.Childs, DBNode{
					Name: xword.Word,
					Type: List,
				})
				n = &n.Childs[len(n.Childs)-1]
			}

			// Ensure List-Key leaf
			found2 := false
			for idx := range n.Childs {
				match := true
				for k, v := range xword.Keys {
					for _, c := range n.Childs[idx].Childs {
						if c.Name == k && !reflect.DeepEqual(c.Value, v) {
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
				for k, v := range xword.Keys {
					tmp := DBNode{
						Name:  k,
						Type:  Leaf,
						Value: v,
					}
					listChilds = append(listChilds, tmp)
				}
				n.Childs = append(n.Childs, DBNode{
					Type:   Container,
					Childs: listChilds,
				})
				n = &n.Childs[len(n.Childs)-1]
			}
		case Leaf:
			switch xword.Dbvaluetype {

			// case yang.Ybits:
			// case yang.Ydecimal64:
			// case yang.Yempty:
			// case yang.Yidentityref:
			// case yang.YinstanceIdentifier:
			// case yang.Yleafref:
			// case yang.Yunion:

			case yang.Yint8:
				v := DBValue{Type: xword.Dbvaluetype}
				if err := v.SetFromString(val); err != nil {
					return nil, errors.Wrap(err, "SetFromString")
				}
				n.Childs = append(n.Childs, DBNode{
					Name:  xword.Word,
					Type:  Leaf,
					Value: v,
				})
			case yang.Yint16:
				v := DBValue{Type: xword.Dbvaluetype}
				if err := v.SetFromString(val); err != nil {
					return nil, errors.Wrap(err, "SetFromString")
				}
				n.Childs = append(n.Childs, DBNode{
					Name:  xword.Word,
					Type:  Leaf,
					Value: v,
				})
			case yang.Yint32:
				v := DBValue{Type: xword.Dbvaluetype}
				if err := v.SetFromString(val); err != nil {
					return nil, errors.Wrap(err, "SetFromString")
				}
				n.Childs = append(n.Childs, DBNode{
					Name:  xword.Word,
					Type:  Leaf,
					Value: v,
				})
			case yang.Yint64:
				v := DBValue{Type: xword.Dbvaluetype}
				if err := v.SetFromString(val); err != nil {
					return nil, errors.Wrap(err, "SetFromString")
				}
				n.Childs = append(n.Childs, DBNode{
					Name:  xword.Word,
					Type:  Leaf,
					Value: v,
				})
			case yang.Yuint8:
				v := DBValue{Type: xword.Dbvaluetype}
				if err := v.SetFromString(val); err != nil {
					return nil, errors.Wrap(err, "SetFromString")
				}
				n.Childs = append(n.Childs, DBNode{
					Name:  xword.Word,
					Type:  Leaf,
					Value: v,
				})
			case yang.Yuint16:
				v := DBValue{Type: xword.Dbvaluetype}
				if err := v.SetFromString(val); err != nil {
					return nil, errors.Wrap(err, "SetFromString")
				}
				n.Childs = append(n.Childs, DBNode{
					Name:  xword.Word,
					Type:  Leaf,
					Value: v,
				})
			case yang.Yuint32:
				v := DBValue{Type: xword.Dbvaluetype}
				if err := v.SetFromString(val); err != nil {
					return nil, errors.Wrap(err, "SetFromString")
				}
				n.Childs = append(n.Childs, DBNode{
					Name:  xword.Word,
					Type:  Leaf,
					Value: v,
				})
			case yang.Yuint64:
				v := DBValue{Type: xword.Dbvaluetype}
				if err := v.SetFromString(val); err != nil {
					return nil, errors.Wrap(err, "SetFromString")
				}
				n.Childs = append(n.Childs, DBNode{
					Name:  xword.Word,
					Type:  Leaf,
					Value: v,
				})
			case yang.Ystring:
				n.Childs = append(n.Childs, DBNode{
					Name: xword.Word,
					Type: Leaf,
					Value: DBValue{
						Type:   yang.Ystring,
						String: val,
					},
				})
			case yang.Ybool:
				bval, err := strconv.ParseBool(val)
				if err != nil {
					return nil, err
				}
				n.Childs = append(n.Childs, DBNode{
					Name: xword.Word,
					Type: Leaf,
					Value: DBValue{
						Type:    yang.Ybool,
						Boolean: bval,
					},
				})

			// TODO(slankdev)
			case yang.Yenum:
				n.Childs = append(n.Childs, DBNode{
					Name: xword.Word,
					Type: Leaf,
					Value: DBValue{
						Type:   yang.Ystring,
						String: val,
					},
				})
			default:
				return nil, fmt.Errorf("%s: unsupported(%s)", util.LINE(), xword.Dbvaluetype)
			}
		case LeafList:
			var tmpNode *DBNode
			for idx := range n.Childs {
				if n.Childs[idx].Name == xword.Word {
					tmpNode = &n.Childs[idx]
					break
				}
			}
			if tmpNode == nil {
				n.Childs = append(n.Childs, DBNode{
					Name: xword.Word,
					Type: LeafList,
				})
				tmpNode = &n.Childs[len(n.Childs)-1]
			}

			arrayvalue := []DBValue{}
			for _, s := range strings.Fields(val) {
				arrayvalue = append(arrayvalue, DBValue{
					Type:   yang.Ystring,
					String: s,
				})
			}
			tmpNode.ArrayValue = arrayvalue
		default:
			return nil, fmt.Errorf("%s: unsupported(%s)", util.LINE(), xword.Dbtype)
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
	for _, xword := range xpath.Words {
		n := new(DBNode)
		n.Name = xword.Word
		n.Type = Container

		if xword.Keys != nil {
			n.Type = List
			n.Childs = []DBNode{
				{Type: Container},
			}
			for k, v := range xword.Keys {
				n.Childs[0].Childs = []DBNode{
					{
						Name:  k,
						Type:  Leaf,
						Value: v,
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

func EnsureListNode(listNode *DBNode, kv map[string]DBValue) *DBNode {
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
			Name:  k,
			Type:  Leaf,
			Value: v,
		}
		newElement.Childs = append(newElement.Childs, n)
	}

	listNode.Childs = append(listNode.Childs, newElement)
	return &listNode.Childs[len(listNode.Childs)-1]
}

func matchChild(root *DBNode, kv map[string]DBValue) bool {
	nMatch := 0
	for idx := range root.Childs {
		child := &root.Childs[idx]
		for k, v := range kv {
			if child.Name == k && reflect.DeepEqual(child.Value, v) {
				nMatch++
			}
		}
	}
	return len(kv) == nMatch
}

func lookupChildIdx(root *DBNode, kv map[string]DBValue) int {
	for idx := range root.Childs {
		child := &root.Childs[idx]
		for idx2 := range child.Childs {
			child2 := &child.Childs[idx2]
			nMatch := 0
			for k, v := range kv {
				if child2.Name == k && reflect.DeepEqual(child2.Value, v) {
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
	// leaf-list
	ArrayValue []DBValue
}

// // type DBValueType string
// type DBValueType yang.TypeKind

// const (
// 	YString      DBValueType = "string"
// 	YInteger     DBValueType = "integer"
// 	YBoolean     DBValueType = "boolean"
// 	YStringArray DBValueType = "stringarray"
// 	YEnum        DBValueType = "enum"
// )

// yang.Yint8
// yang.Yint16
// yang.Yint32
// yang.Yint64
// yang.Yuint8
// yang.Yuint16
// yang.Yuint32
// yang.Yuint64
// yang.Ybinary
// yang.Ybits
// yang.Ybool
// yang.Ydecimal64
// yang.Yempty
// yang.Yenum
// yang.Yidentityref
// yang.YinstanceIdentifier
// yang.Yleafref
// yang.Ystring
// yang.Yunion

type DBValue struct {
	Type yang.TypeKind

	// Union
	Int8    int8
	Int16   int16
	Int32   int32
	Int64   int64
	Uint8   uint8
	Uint16  uint16
	Uint32  uint32
	Uint64  uint64
	String  string
	Boolean bool
}

func (v *DBValue) ToString() string {
	switch v.Type {
	case yang.Ystring:
		return v.String
	case yang.Yint8:
		return fmt.Sprintf("%d", v.Int8)
	case yang.Yint16:
		return fmt.Sprintf("%d", v.Int16)
	case yang.Yint32:
		return fmt.Sprintf("%d", v.Int32)
	case yang.Yint64:
		return fmt.Sprintf("%d", v.Int64)
	case yang.Yuint8:
		return fmt.Sprintf("%d", v.Uint8)
	case yang.Yuint16:
		return fmt.Sprintf("%d", v.Uint16)
	case yang.Yuint32:
		return fmt.Sprintf("%d", v.Uint32)
	case yang.Yuint64:
		return fmt.Sprintf("%d", v.Uint64)
	case yang.Ybool:
		return fmt.Sprintf("%v", v.Boolean)
	// case yang.Ybinary:
	// case yang.Ybits:
	// case yang.Ydecimal64:
	// case yang.Yempty:
	// case yang.Yenum:
	// case yang.Yidentityref:
	// case yang.YinstanceIdentifier:
	// case yang.Yleafref:
	// case yang.Yunion:
	default:
		panic(fmt.Sprintf("OKASHI (%s)", v.Type))
	}
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
		array := []string{}
		for _, a := range n.ArrayValue {
			array = append(array, a.String)
		}
		return array
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
		keys := util.GetSortedKeys(g)
		for _, k := range keys {
			v := g[k]
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
			n.ArrayValue = v
		} else {
			n.Type = List
			n.Childs = append(n.Childs, childCandidates...)
		}
	case bool:
		n.Type = Leaf
		n.Value = DBValue{
			Type:    yang.Ybool,
			Boolean: g,
		}
	// case int:
	// 	n.Type = Leaf
	// 	n.Value = DBValue{
	// 		Type:  yang.Yint32,
	// 		Int32: int32(g),
	// 	}
	case int8:
		n.Type = Leaf
		n.Value = DBValue{
			Type: yang.Yint8,
			Int8: int8(g),
		}
	case int16:
		n.Type = Leaf
		n.Value = DBValue{
			Type:  yang.Yint16,
			Int16: int16(g),
		}
	case int32:
		n.Type = Leaf
		n.Value = DBValue{
			Type:  yang.Yint32,
			Int32: int32(g),
		}
	case int64:
		n.Type = Leaf
		n.Value = DBValue{
			Type:  yang.Yint64,
			Int64: int64(g),
		}
	case uint8:
		n.Type = Leaf
		n.Value = DBValue{
			Type:  yang.Yuint8,
			Uint8: uint8(g),
		}
	case uint16:
		n.Type = Leaf
		n.Value = DBValue{
			Type:   yang.Yuint16,
			Uint16: uint16(g),
		}
	case uint32:
		n.Type = Leaf
		n.Value = DBValue{
			Type:   yang.Yuint32,
			Uint32: uint32(g),
		}
	case uint64:
		n.Type = Leaf
		n.Value = DBValue{
			Type:   yang.Yuint64,
			Uint64: uint64(g),
		}
	case string:
		n.Type = Leaf
		n.Value = DBValue{
			Type:   yang.Ystring,
			String: g,
		}
	case []string:
		a := []DBValue{}
		for _, s := range g {
			a = append(a, DBValue{
				Type:   yang.Ystring,
				String: s,
			})
		}
		n.Type = LeafList
		n.ArrayValue = a

	// TODO(slankdev)
	case float64:
		n.Type = Leaf
		n.Value = DBValue{
			Type:  yang.Yint32,
			Int32: int32(g),
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
	case yang.Yint8:
		return v.Int8
	case yang.Yint16:
		return v.Int16
	case yang.Yint32:
		return v.Int32
	case yang.Yint64:
		return v.Int64
	case yang.Yuint8:
		return v.Uint8
	case yang.Yuint16:
		return v.Uint16
	case yang.Yuint32:
		return v.Uint32
	case yang.Yuint64:
		return v.Uint64
	case yang.Ybool:
		return v.Boolean
	case yang.Ystring:
		return v.String
	default:
		panic(fmt.Sprintf("ASSERT(%s)", v.Type))
	}
}

func (v *DBValue) SetFromString(s string) error {
	switch v.Type {
	case yang.Yint8:
		ival, err := strconv.ParseInt(s, 10, 8)
		if err != nil {
			return errors.Wrap(err, "strconv.ParseInt(s,10,8)")
		}
		v.Int8 = int8(ival)
	case yang.Yint16:
		ival, err := strconv.ParseInt(s, 10, 16)
		if err != nil {
			return errors.Wrap(err, "strconv.ParseInt(s,10,16)")
		}
		v.Int16 = int16(ival)
	case yang.Yint32:
		ival, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return errors.Wrap(err, "strconv.ParseInt(s,10,32)")
		}
		v.Int32 = int32(ival)
	case yang.Yint64:
		ival, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return errors.Wrap(err, "strconv.ParseInt(s,10,64)")
		}
		v.Int64 = int64(ival)
	case yang.Yuint8:
		ival, err := strconv.ParseUint(s, 10, 8)
		if err != nil {
			return errors.Wrap(err, "strconv.ParseUint(s,10,8)")
		}
		v.Uint8 = uint8(ival)
	case yang.Yuint16:
		ival, err := strconv.ParseUint(s, 10, 16)
		if err != nil {
			return errors.Wrap(err, "strconv.ParseUint(s,10,16)")
		}
		v.Uint16 = uint16(ival)
	case yang.Yuint32:
		ival, err := strconv.ParseUint(s, 10, 32)
		if err != nil {
			return errors.Wrap(err, "strconv.ParseUint(s,10,32)")
		}
		v.Uint32 = uint32(ival)
	case yang.Yuint64:
		ival, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return errors.Wrap(err, "strconv.ParseUint(s,10,64)")
		}
		v.Uint64 = uint64(ival)
	case yang.Ystring:
		v.String = s
	case yang.Ybool:
		bval, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		v.Boolean = bval

	// case yang.Ybits:
	// case yang.Ydecimal64:
	// case yang.Yempty:
	// case yang.Yenum:
	// case yang.Yidentityref:
	// case yang.YinstanceIdentifier:
	// case yang.Yleafref:
	// case yang.Yunion:
	default:
		panic(fmt.Sprintf("OKASHI (%s)", v.Type))
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

func squashListToLeafList(items []DBNode) ([]DBValue, error) {
	typesMap := map[yang.TypeKind]bool{}
	for _, item := range items {
		typesMap[item.Value.Type] = true
	}
	types := []yang.TypeKind{}
	for key, val := range typesMap {
		if val {
			types = append(types, key)
		}
	}
	if len(types) != 1 {
		return []DBValue{}, fmt.Errorf("invalid list items (%+v)", types)
	}

	switch types[0] {
	// TODO(slankdev): implemente me
	// case YInteger:
	// case YBoolean:

	case yang.Ystring:
		a := []DBValue{}
		for _, item := range items {
			a = append(a, DBValue{
				Type:   yang.Ystring,
				String: item.Value.String,
			})
		}
		return a, nil
	default:
		panic(fmt.Sprintf("OKASHII %s", types[0]))
	}
}
