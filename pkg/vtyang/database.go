package vtyang

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/openconfig/goyang/pkg/indent"
)

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
	ListChilds [][]DBNode
	Value      DBValue
}

func (n *DBNode) String() string {
	buf := new(bytes.Buffer)
	n.Write(buf)
	return buf.String()
}

func (n *DBNode) Write(w io.Writer) {
	switch n.Type {
	case Container:
		if n.Name != "" {
			fmt.Fprintf(w, "\"%s\": ", n.Name)
		}
		fmt.Fprintf(w, "{\n")
		for _, child := range n.Childs {
			child.Write(indent.NewWriter(w, "  "))
		}
		fmt.Fprintf(w, "}\n")
	case List:
		fmt.Fprintf(w, "\"%s\": [\n", n.Name)
		for _, child := range n.Childs {
			child.Write(indent.NewWriter(w, "  "))
		}
		fmt.Fprintf(w, "]\n")
	case Leaf:
		fmt.Fprintf(w, "\"%s\": %s\n", n.Name, n.Value.ToJsonValue())
	}
}

func (n *DBNode) JSONString() string {
	m := n.ToMap()
	return js(&m)
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
	default:
		panic(fmt.Sprintf("ASSERT(%s)", n.Type))
	}
	return m
}

func js(i interface{}) string {
	b, err := json.Marshal(&i)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return "{}"
	}
	var out bytes.Buffer
	if err = json.Indent(&out, b, "", "  "); err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return "{}"
	}
	return out.String()
}

type DBValueType string

const (
	YString  DBValueType = "string"
	YInteger DBValueType = "integer"
	YBoolean DBValueType = "boolean"
)

type DBValue struct {
	Type DBValueType

	// Union
	Integer int
	String  string
	Boolean bool
}

func (v DBValue) ToValue() interface{} {
	switch v.Type {
	case YInteger:
		return v.Integer
	case YBoolean:
		return v.Boolean
	case YString:
		return v.String
	default:
		panic(fmt.Sprintf("ASSERT(%s)", v.Type))
	}
}

func (v DBValue) ToJsonValue() string {
	switch v.Type {
	case YInteger:
		return fmt.Sprintf("%d", v.Integer)
	case YBoolean:
		return strconv.FormatBool(v.Boolean)
	case YString:
		return fmt.Sprintf("\"%s\"", v.String)
	}
	panic(fmt.Sprintf("unsupported \"%s\"", v.Type))
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

type DB struct {
	active bool
	root   DBNode
}

func (dbm *DatabaseManager) GetNode(mod string, xpath XPath) (*DBNode, error) {
	n := &dbm.db.root
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

func (dbm *DatabaseManager) DeleteNode(mod string, xpath XPath) error {
	n := &dbm.db.root
	xwords := xpath.words

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

func (dbm *DatabaseManager) SetNode(mod string, xpath XPath, val string) (
	*DBNode, error) {

	n := &dbm.db.root
	xwords := xpath.words

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
				//pp.Println(newnode.Value)
			}

			n.Childs = append(n.Childs, newnode)
			n = &n.Childs[len(n.Childs)-1]
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

func MergeDBNode(org, new DBNode) (DBNode, error) {
	result := DBNode{
		Name: "<root>",
		Type: Container,
		Childs: []DBNode{
			{
				Name: "users",
				Type: Container,
				Childs: []DBNode{
					{
						Name: "user",
						Type: List,
						Childs: []DBNode{
							{
								Type: Container,
								Childs: []DBNode{
									{
										Name: "name",
										Type: Leaf,
										Value: DBValue{
											Type:   YString,
											String: "hoge",
										},
									},
								},
							},
							{
								Type: Container,
								Childs: []DBNode{
									{
										Name: "name",
										Type: Leaf,
										Value: DBValue{
											Type:   YString,
											String: "fuga",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return result, nil
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
