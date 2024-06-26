package vtyang

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/nsf/jsondiff"
	"github.com/openconfig/goyang/pkg/yang"
	"github.com/pkg/errors"

	"github.com/slankdev/vtyang/pkg/util"
)

type DatabaseManager struct {
	// root
	root DBNode
	// candidateRoot is the top of candidate config
	candidateRoot *DBNode
}

func NewDatabaseManager() *DatabaseManager {
	m := DatabaseManager{}
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

func yangModuleDumpEntries() []*yang.Entry {
	entries := []*yang.Entry{}
	for _, m := range yangmodules.Modules {
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
									if child3.Name == k && reflect.DeepEqual(child3.Value, v.Value) {
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
										Value: v.Value,
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
						if c.Name == k && !reflect.DeepEqual(c.Value, v.Value) {
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
						Value: v.Value,
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
			v := DBValue{
				Type:      xword.Dbvaluetype,
				UnionType: xword.Dbuniontype,
			}
			if err := v.SetFromStringWithType(val, xword); err != nil {
				return nil, errors.Wrap(err, "SetFromStringWithType")
			}
			n.Childs = append(n.Childs, DBNode{
				Name:  xword.Word,
				Type:  Leaf,
				Value: v,
			})
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
						Value: v.Value,
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

func EnsureListNode(listNode *DBNode, kv map[string]XWordKey) *DBNode {
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
			Value: v.Value,
		}
		newElement.Childs = append(newElement.Childs, n)
	}

	listNode.Childs = append(listNode.Childs, newElement)
	return &listNode.Childs[len(listNode.Childs)-1]
}

func matchChild(root *DBNode, kv map[string]XWordKey) bool {
	nMatch := 0
	for idx := range root.Childs {
		child := &root.Childs[idx]
		for k, v := range kv {
			if child.Name == k && reflect.DeepEqual(child.Value, v.Value) {
				nMatch++
			}
		}
	}
	return len(kv) == nMatch
}

func lookupChildIdx(root *DBNode, kv map[string]XWordKey) int {
	for idx := range root.Childs {
		child := &root.Childs[idx]
		for idx2 := range child.Childs {
			child2 := &child.Childs[idx2]
			nMatch := 0
			for k, v := range kv {
				if child2.Name == k && reflect.DeepEqual(child2.Value, v.Value) {
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

type DBNodeType string

const (
	Container DBNodeType = "container"
	List      DBNodeType = "list"
	Leaf      DBNodeType = "leaf"
	LeafList  DBNodeType = "leaf-list"
)

type DBNode struct {
	Name       string
	Type       DBNodeType
	Childs     []DBNode
	Value      DBValue
	ArrayValue []DBValue
}

type DBValue struct {
	Type      yang.TypeKind
	UnionType yang.TypeKind `json:"UnionType,omitempty"`
	Int8      int8
	Int16     int16
	Int32     int32
	Int64     int64
	Uint8     uint8
	Uint16    uint16
	Uint32    uint32
	Uint64    uint64
	String    string
	Boolean   bool
	Decimal64 float64
}

func (v *DBValue) ToAbsoluteNumber() (uint64, bool, error) {
	switch v.Type {
	case yang.Ydecimal64:
		return uint64(math.Abs(float64(v.Decimal64))), v.Decimal64 < 0, nil
	case yang.Yint8:
		return uint64(math.Abs(float64(v.Int8))), v.Int8 < 0, nil
	case yang.Yint16:
		return uint64(math.Abs(float64(v.Int16))), v.Int16 < 0, nil
	case yang.Yint32:
		return uint64(math.Abs(float64(v.Int32))), v.Int32 < 0, nil
	case yang.Yint64:
		b := big.NewInt(v.Int64)
		return b.Abs(b).Uint64(), v.Int64 < 0, nil

	case yang.Yuint8:
		return uint64(v.Uint8), false, nil
	case yang.Yuint16:
		return uint64(v.Uint16), false, nil
	case yang.Yuint32:
		return uint64(v.Uint32), false, nil
	case yang.Yuint64:
		return (v.Uint64), false, nil
	default:
		panic(fmt.Sprintf("OKASHI (%s)", v.Type))
	}
}

func (v *DBValue) ToYangNumber() (*yang.Number, error) {
	uv, neg, err := v.ToAbsoluteNumber()
	if err != nil {
		return nil, errors.Wrap(err, "ToAbsoluteNumber")
	}
	n := yang.Number{
		Value:    uv,
		Negative: neg,
	}
	return &n, nil
}

func (v *DBValue) ToString() string {
	switch v.Type {
	case yang.Ystring,
		yang.Yidentityref,
		yang.Yleafref,
		yang.Yenum:
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
	case yang.Ydecimal64:
		return fmt.Sprintf("%f", v.Decimal64)
	case yang.Yunion:
		vv := *v
		vv.Type = vv.UnionType
		return vv.ToString()
	// case yang.Ybinary:
	// case yang.Ybits:
	// case yang.Yempty:
	// case yang.Yenum:
	// case yang.YinstanceIdentifier:
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
	if err := os.WriteFile(filename, []byte(s), 0644); err != nil {
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
	case float64:
		n.Type = Leaf
		n.Value = DBValue{
			Type:      yang.Ydecimal64,
			Decimal64: float64(g),
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
	case yang.Ydecimal64:
		return v.Decimal64
	case yang.Yleafref:
		return v.String
	case yang.Yidentityref:
		return v.String
	case yang.Yenum:
		return v.String
	case yang.Yunion:
		vv := v
		vv.Type = vv.UnionType
		vv.UnionType = yang.Ynone
		return vv.ToValue()
	default:
		panic(fmt.Sprintf("ASSERT(%s)", v.Type))
	}
}

func (v *DBValue) SetFromStringWithType(s string, xword XWord) error {
	if v.Type == yang.Yunion && v.UnionType == yang.Ynone {
		validated := false
		unionTypes := resolveUnionTypes(xword.ytype.Type)
		for _, ytype := range unionTypes {
			switch ytype.Kind {
			case yang.Ystring:
				if err := validateStringValue(s, ytype); err == nil {
					v.UnionType = ytype.Kind
					validated = true
				}
			case yang.Yenum:
				if err := validateEnumValue(s, ytype); err == nil {
					v.UnionType = ytype.Kind
					validated = true
				}
			case
				yang.Yint8,
				yang.Yint16,
				yang.Yint32,
				yang.Yint64,
				yang.Yuint8,
				yang.Yuint16,
				yang.Yuint32,
				yang.Yuint64,
				yang.Ydecimal64:
				if err := validateNumberValue(s, ytype); err == nil {
					v.UnionType = ytype.Kind
					validated = true
				}
			default:
				panic(fmt.Sprintf("PANIC %s", ytype.Kind.String()))
			}
			if validated {
				break
			}
		}
		if !validated {
			return errors.Errorf("union not validated")
		}
	}
	return v.SetFromString(s)
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
	case yang.Ydecimal64:
		ival, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return errors.Wrap(err, "strconv.ParseFloat(s,64)")
		}
		v.Decimal64 = float64(ival)
	case yang.Ybool:
		bval, err := strconv.ParseBool(s)
		if err != nil {
			return errors.Wrap(err, "strconv.ParseBool")
		}
		v.Boolean = bval
	case yang.Yunion:
		vv := DBValue{}
		vv.Type = v.UnionType
		if err := vv.SetFromString(s); err != nil {
			return errors.Wrap(err, "vv.SetFromString")
		}
		*v = vv
		v.UnionType = v.Type
		v.Type = v.UnionType

	// TODO(slankdev)
	case yang.Ystring,
		yang.Yenum,
		yang.Yleafref,
		yang.Yidentityref:
		v.String = s

	// case yang.Ybits:
	// case yang.Yempty:
	// case yang.YinstanceIdentifier:
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
		for _, e := range root.Dir {
			for _, child := range n.Childs {
				if child.Name == e.Name {
					childs = append(childs, *filterDbWithModuleImpl(&child, e))
				}
			}
		}
	case root.IsLeafList():
		if root.Name == n.Name {
			childs = append(childs, *n)
		}
	case root.IsLeaf():
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
