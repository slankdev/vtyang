package vtyang

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/nsf/jsondiff"
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
	Boolean bool
}

func (n *DBNode) DeepCopy() *DBNode {
	m := n.ToMap()
	copy, _ := Interface2DBNode(m)
	return copy
}

func (n *DBNode) String() string {
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
	raw, err := ioutil.ReadFile(filename)
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
		for _, v := range g {
			child, err := Interface2DBNode(v)
			if err != nil {
				return nil, err
			}
			n.Type = List
			n.Childs = append(n.Childs, *child)
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
