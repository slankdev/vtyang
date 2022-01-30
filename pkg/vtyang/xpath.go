package vtyang

import (
	"fmt"
	"strings"

	"github.com/openconfig/goyang/pkg/yang"
)

type XWord struct {
	word string
	keys map[string]string

	dbtype      DBNodeType
	dbvaluetype DBValueType
}

type XPath struct {
	words []XWord
}

func NewXPath(s string) (XPath, error) {
	xp := XPath{}
	if err := ParseXPath(&xp, s); err != nil {
		return XPath{}, err
	}
	return xp, nil
}

func NewXPathOrDie(s string) XPath {
	xp, err := NewXPath(s)
	ErrorOnDie(err)
	return xp
}

func ParseXPath(xpath *XPath, s string) error {
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
			xword.keys = map[string]string{}
			xword.keys[k] = v
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
				s = fmt.Sprintf("%s['%s'='%s']", s, k, v)
			}
		}
	}
	return s
}

func ParseXPathArgs(args []string, setmode bool) (XPath, string, error) {
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
			//fmt.Printf("vs %s %s\n", n, words[0])
			if n == words[0] {
				foundNode = module.Dir[n]
				break
			}
		}
		if foundNode == nil {
			return XPath{}, "", fmt.Errorf("entry %s is not found", words[0])
		}

		argumentExist := false
		switch {
		case foundNode.IsContainer():
			xword.dbtype = Container
		case foundNode.IsLeaf():
			if setmode {
				if len(words) < 2 {
					return XPath{}, "", fmt.Errorf("invalid args len")
				}
				valueStr = words[1]
				argumentExist = true
			}
			xword.dbtype = Leaf
			xword.dbvaluetype = YangTypeKind2YType(foundNode.Type.Kind)
		case foundNode.IsList():
			if len(words) < 2 {
				return XPath{}, "", fmt.Errorf("invalid args len")
			}
			xword.dbtype = List
			xword.keys = map[string]string{}
			xword.keys[foundNode.Key] = words[1]
			argumentExist = true
		default:
			panic("ASSERT")
		}

		xpath.words = append(xpath.words, xword)
		words = words[1:]
		if argumentExist {
			words = words[1:]
		}
		module = foundNode
	}

	return xpath, valueStr, nil
}
