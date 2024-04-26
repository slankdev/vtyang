package vtyang

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/k0kubun/pp"
	"github.com/openconfig/goyang/pkg/yang"
)

type XWord struct {
	Word        string
	Keys        map[string]DBValue
	Dbtype      DBNodeType
	Dbvaluetype DBValueType
}

type XPath struct {
	Words []XWord
}

func NewXPath(dbm *DatabaseManager, s string) (XPath, error) {
	xp := XPath{}
	if err := ParseXPath(dbm, &xp, s); err != nil {
		return XPath{}, err
	}
	return xp, nil
}

func NewXPathOrDie(dbm *DatabaseManager, s string) XPath {
	xp, err := NewXPath(dbm, s)
	ErrorOnDie(err)
	return xp
}

func ParseXPath(dbm *DatabaseManager, xpath *XPath, s string) error {
	words := strings.FieldsFunc(s, func(c rune) bool {
		return c == '/'
	})

	module := &yang.Entry{}
	module.Dir = map[string]*yang.Entry{}
	for _, ent := range dbm.DumpEntries() {
		module.Dir[ent.Name] = ent
	}

	for len(words) != 0 {
		xword := XWord{Word: name(words[0])}
		if hasKV(words[0]) {
			k := key(words[0])
			v := val(words[0])
			xword.Keys = map[string]DBValue{}
			xword.Keys[k] = DBValue{
				Type:   YString,
				String: v,
			}
		}

		var foundNode *yang.Entry = nil
		for n := range module.Dir {
			if n == name(words[0]) {
				foundNode = module.Dir[n]
				break
			}
		}
		if foundNode == nil {
			return fmt.Errorf("entry %s is not found", name(words[0]))
		}

		switch {
		case foundNode.IsContainer():
			xword.Dbtype = Container
		case foundNode.IsLeaf():
			xword.Dbtype = Leaf
			xword.Dbvaluetype = YangTypeKind2YType(foundNode.Type.Kind)
		case foundNode.IsList():
			xword.Dbtype = List
		default:
			panic("ASSERT")
		}

		xpath.Words = append(xpath.Words, xword)
		words = words[1:]
		module = foundNode
	}

	return nil
}

func (x XPath) String() string {
	s := ""
	for _, w := range x.Words {
		s = fmt.Sprintf("%s/%s", s, w.Word)
		if w.Keys != nil {
			for k, v := range w.Keys {
				s = fmt.Sprintf("%s['%s'='%s']", s, k, v.ToString())
			}
		}
	}
	return s
}

func ParseXPathArgs(dbm *DatabaseManager, args []string, setmode bool) (XPath, string, error) {
	var xpath XPath
	var val string
	var err error
	for _, ent := range dbm.DumpEntries() {
		module := &yang.Entry{}
		module.Dir = map[string]*yang.Entry{}
		module.Dir[ent.Name] = ent
		xpath, val, err = ParseXPathArgsImpl(module, args, setmode)
		if err == nil {
			break
		}
	}
	if err != nil {
		return XPath{}, "", err
	}
	return xpath, val, nil
}

func ParseXPathArgsImpl(module *yang.Entry, args []string, setmode bool) (XPath, string, error) {
	words := args
	xpath := XPath{}
	valueStr := ""
	for len(words) != 0 {
		xword := XWord{Word: words[0]}

		var foundNode *yang.Entry = nil
		for n := range module.Dir {
			if n == words[0] {
				foundNode = module.Dir[n]
				break
			}
		}
		if foundNode == nil {
			return XPath{}, "", fmt.Errorf("entry %s is not found", words[0])
		}

		argumentCount := 1
		argumentExist := false
		switch {
		case foundNode.IsContainer():
			xword.Dbtype = Container
		case foundNode.IsLeaf():
			if setmode {
				if len(words) < 2 {
					return XPath{}, "", fmt.Errorf("invalid args len")
				}
				valueStr = words[argumentCount]
				argumentExist = true
			}
			xword.Dbtype = Leaf
			xword.Dbvaluetype = YangTypeKind2YType(foundNode.Type.Kind)
		case foundNode.IsList():
			if len(words) < 2 {
				return XPath{}, "", fmt.Errorf("invalid args len")
			}
			xword.Dbtype = List
			xword.Keys = map[string]DBValue{}
			for _, w := range strings.Fields(foundNode.Key) {
				var keyLeafNode *yang.Entry
				for _, ee := range foundNode.Dir {
					if ee.Name == w {
						keyLeafNode = ee
						break
					}
				}
				tmpStr := words[argumentCount]
				switch keyLeafNode.Type.Name {
				case "string":
					xword.Keys[w] = DBValue{
						Type:   YString,
						String: tmpStr,
					}
				case "vrf-ref":
					xword.Keys[w] = DBValue{
						Type:   YString,
						String: tmpStr,
					}
				case "ip-prefix":
					xword.Keys[w] = DBValue{
						Type:   YString,
						String: tmpStr,
					}
				case "uint32":
					intval, err := strconv.ParseInt(tmpStr, 10, 32)
					if err != nil {
						return XPath{}, "", err
					}
					xword.Keys[w] = DBValue{
						Type:    YInteger,
						Integer: int(intval),
					}
				case "enumeration":
					pp.Println(keyLeafNode.Type)
					panic(fmt.Sprintf("NOT IMPLE (%+v)", keyLeafNode.Type.Name))
				default:
					panic(fmt.Sprintf("OKASHII (%+v)", keyLeafNode.Type.Name))
				}
				argumentCount++
			}
			argumentCount--
			argumentExist = true
		case foundNode.IsLeafList():
			if setmode {
				if len(words) < 2 {
					return XPath{}, "", fmt.Errorf("invalid args len")
				}
				vals := []string{}
				for argumentCount < len(words) {
					vals = append(vals, words[argumentCount])
					argumentCount++
				}
				argumentCount--
				valueStr = strings.Join(vals, " ")
				argumentExist = true
			}
			xword.Dbtype = LeafList
			xword.Dbvaluetype = YangTypeKind2YType(foundNode.Type.Kind)
			//pp.Println(valueStr)
		default:
			panic("ASSERT")
		}

		xpath.Words = append(xpath.Words, xword)
		words = words[1:]
		if argumentExist {
			words = words[argumentCount:]
		}
		module = foundNode
	}

	return xpath, valueStr, nil
}
