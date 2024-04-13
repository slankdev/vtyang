package vtyang

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/openconfig/goyang/pkg/yang"
)

type XWord struct {
	word string
	keys map[string]DBValue

	dbtype      DBNodeType
	dbvaluetype DBValueType
}

type XPath struct {
	words []XWord
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
		xword := XWord{word: name(words[0])}
		if hasKV(words[0]) {
			k := key(words[0])
			v := val(words[0])
			xword.keys = map[string]DBValue{}
			xword.keys[k] = DBValue{
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
			xword.dbtype = Container
		case foundNode.IsLeaf():
			xword.dbtype = Leaf
			xword.dbvaluetype = YangTypeKind2YType(foundNode.Type.Kind)
		case foundNode.IsList():
			xword.dbtype = List
		default:
			panic("ASSERT")
		}

		xpath.words = append(xpath.words, xword)
		words = words[1:]
		module = foundNode
	}

	return nil
}

func (x XPath) String() string {
	s := ""
	for _, w := range x.words {
		s = fmt.Sprintf("%s/%s", s, w.word)
		if w.keys != nil {
			for k, v := range w.keys {
				s = fmt.Sprintf("%s['%s'='%s']", s, k, v.ToString())
			}
		}
	}
	return s
}

func ParseXPathArgs(dbm *DatabaseManager, args []string, setmode bool) (XPath, string, error) {
	module := &yang.Entry{}
	module.Dir = map[string]*yang.Entry{}
	for _, ent := range dbm.DumpEntries() {
		module.Dir[ent.Name] = ent
	}

	words := args
	xpath := XPath{}
	valueStr := ""
	for len(words) != 0 {
		xword := XWord{word: words[0]}

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
			xword.dbtype = Container
		case foundNode.IsLeaf():
			if setmode {
				if len(words) < 2 {
					return XPath{}, "", fmt.Errorf("invalid args len")
				}
				valueStr = words[argumentCount]
				argumentExist = true
			}
			xword.dbtype = Leaf
			xword.dbvaluetype = YangTypeKind2YType(foundNode.Type.Kind)
		case foundNode.IsList():
			if len(words) < 2 {
				return XPath{}, "", fmt.Errorf("invalid args len")
			}
			xword.dbtype = List
			xword.keys = map[string]DBValue{}
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
					xword.keys[w] = DBValue{
						Type:   YString,
						String: tmpStr,
					}
				case "vrf-ref":
					xword.keys[w] = DBValue{
						Type:   YString,
						String: tmpStr,
					}
				case "ip-prefix":
					xword.keys[w] = DBValue{
						Type:   YString,
						String: tmpStr,
					}
				case "uint32":
					intval, err := strconv.ParseInt(tmpStr, 10, 32)
					if err != nil {
						return XPath{}, "", err
					}
					xword.keys[w] = DBValue{
						Type:    YInteger,
						Integer: int(intval),
					}
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
			xword.dbtype = LeafList
			xword.dbvaluetype = YangTypeKind2YType(foundNode.Type.Kind)
			//pp.Println(valueStr)
		default:
			panic("ASSERT")
		}

		xpath.words = append(xpath.words, xword)
		words = words[1:]
		if argumentExist {
			words = words[argumentCount:]
		}
		module = foundNode
	}

	return xpath, valueStr, nil
}
