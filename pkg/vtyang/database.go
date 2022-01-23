package vtyang

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/openconfig/goyang/pkg/yang"
	"github.com/slankdev/vtyang/pkg/util"
)

type DatabaseManager struct {
	modules *yang.Modules
	root    DBNode

	// candidateRoot is the top of candidate config
	candidateRoot *DBNode
}

var dbm *DatabaseManager

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
	root, err := ReadFromJsonFile(config.GlobalOptDBPath)
	if err != nil {
		return err
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

func (m *DatabaseManager) YangEntries() []*yang.Entry {
	return dbm.DumpEntries()
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
		if n.Type != Container {
			return nil, fmt.Errorf("%s: unsupported(%s)", util.LINE(), n.Type)
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
