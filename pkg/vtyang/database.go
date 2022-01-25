package vtyang

import (
	"strings"

	"github.com/slankdev/vtyang/pkg/util"
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
	Boolean string
}

type DB struct {
	active bool
	root   DBNode
}

func (dbm DatabaseManager) GetNode(mod, xpath string) (*DBNode, error) {
	root := &dbm.db.root
	words := strings.FieldsFunc(xpath, func(c rune) bool {
		return c == '/'
	})

	n, err := dbm.DigContainer(words, root)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (dbm DatabaseManager) DigContainer(words []string, n *DBNode) (*DBNode, error) {
	if len(words) == 0 {
		return n, nil
	}
	// fmt.Printf("DEBUG %s(%s), %+v\n", n.Name, n.Type, words[0])

	name := func(s string) string {
		ret := util.SplitMultiSep(s, []string{"'", "[", "]", "="})
		return ret[0]
	}
	key := func(s string) string {
		ret := util.SplitMultiSep(s, []string{"'", "[", "]", "="})
		if len(ret) != 3 {
			panic(s)
		}
		return ret[1]
	}
	val := func(s string) string {
		ret := util.SplitMultiSep(s, []string{"'", "[", "]", "="})
		if len(ret) != 3 {
			panic(s)
		}
		return ret[2]
	}

	switch n.Type {
	case Container:
		for idx := range n.Childs {
			child := &n.Childs[idx]
			if n.Childs[idx].Name == name(words[0]) {
				switch n.Childs[idx].Type {
				case Leaf:
					fallthrough
				case Container:
					return dbm.DigContainer(words[1:], &n.Childs[idx])
				case List:
					k := key(words[0])
					v := val(words[0])
					for idx2 := range child.Childs {
						child2 := &child.Childs[idx2]
						for idx3 := range child.Childs[idx2].Childs {
							child3 := &child.Childs[idx2].Childs[idx3]
							if child3.Name == k && child3.Value.String == v {
								return dbm.DigContainer(words[1:], child2)
							}
						}
					}
				default:
					panic("UNSUPPORTED")
				}
			}
		}
	case List:
		panic("UNSUPPORTED")

	default:
		panic("UNSUPPORTED")
	}

	// println("NOTFOUND!!")
	return nil, nil
}
