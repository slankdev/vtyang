package vtyang

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/openconfig/goyang/pkg/yang"
	"github.com/pkg/errors"
)

type XWord struct {
	Word        string
	Keys        map[string]DBValue
	Dbtype      DBNodeType
	Dbvaluetype yang.TypeKind
	Dbuniontype yang.TypeKind `json:"Dbuniontype,omitempty"`
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
				Type:   yang.Ystring,
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
			xword.Dbvaluetype = foundNode.Type.Kind
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
			e := module.Dir[n]
			switch {
			case e.IsChoice():
				for _, ee := range e.Dir {
					switch {
					case ee.IsCase():
						for _, eee := range ee.Dir {
							if eee.Name == words[0] {
								foundNode = eee
							}
						}
					default:
						panic("OKASHII")
					}
				}
			default:
				if n == words[0] {
					foundNode = module.Dir[n]
				}
			}
			if foundNode != nil {
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
			xword.Dbvaluetype = foundNode.Type.Kind

			// Additional Validation for Union
			switch foundNode.Type.Kind {
			case yang.Yunion:
				validated := false
				ytypes := resolveUnionTypes(foundNode.Type.Type)
				for _, ytype := range ytypes {
					switch ytype.Kind {
					case yang.Ystring:
						if err := validateStringValue(valueStr, ytype); err == nil {
							xword.Dbuniontype = ytype.Kind
							validated = true
						}
					case yang.Yenum:
						if err := validateEnumValue(valueStr, ytype); err == nil {
							xword.Dbuniontype = ytype.Kind
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
						if err := validateNumberValue(valueStr, ytype); err == nil {
							xword.Dbuniontype = ytype.Kind
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
					return XPath{}, "", errors.Errorf("union not validated")
				}
			}

			// Additional Validation for String
			switch foundNode.Type.Kind {
			case yang.Ystring:
				if err := validateStringValue(valueStr, foundNode.Type); err != nil {
					return XPath{}, "", errors.Wrap(err, "validateStringValue")
				}
			}

			// Additional Validation for Enum
			switch foundNode.Type.Kind {
			case yang.Yenum:
				if err := validateEnumValue(valueStr, foundNode.Type); err != nil {
					return XPath{}, "", errors.Wrap(err, "validateEnumValue")
				}
			}

			// Additional Validatation for Number-types
			// - range (A)..(B)
			switch foundNode.Type.Kind {
			case yang.Yint8, yang.Yint16, yang.Yint32, yang.Yint64,
				yang.Yuint8, yang.Yuint16, yang.Yuint32, yang.Yuint64,
				yang.Ydecimal64:
				if err := validateNumberValue(valueStr, foundNode.Type); err != nil {
					return XPath{}, "", errors.Wrap(err, "validateNumberValue")
				}
			}

			// Additional Validation for String-types
			// - length
			// - pattern
			// - pattern(invert-match)
			// TODO(slankdev): IMPLEMENT ME

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

				// Check union type
				tmpStr := words[argumentCount]
				unionType := yang.Ynone
				if keyLeafNode.Type.Kind == yang.Yunion {
					validated := false
					ytypes := resolveUnionTypes(keyLeafNode.Type.Type)
					for _, ytype := range ytypes {
						switch ytype.Kind {
						case yang.Ystring:
							if err := validateStringValue(valueStr, ytype); err == nil {
								unionType = ytype.Kind
								validated = true
							}
						case yang.Yenum:
							if err := validateEnumValue(valueStr, ytype); err == nil {
								unionType = ytype.Kind
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
							if err := validateNumberValue(valueStr, ytype); err == nil {
								unionType = ytype.Kind
								validated = true
							}
						default:
							panic(fmt.Sprintf("PANIC %s", ytype.Kind.String()))
						}
					}
					if !validated {
						return XPath{}, "", errors.Errorf("union not validated")
					}
				}

				// Prase from string
				v := DBValue{
					Type:      keyLeafNode.Type.Kind,
					UnionType: unionType,
				}
				if err := v.SetFromString(tmpStr); err != nil {
					return XPath{}, "", errors.Wrap(err, "SetFromString")
				}
				xword.Keys[w] = v
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
			xword.Dbvaluetype = foundNode.Type.Kind
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

func validateStringValue(valueStr string, yangType *yang.YangType) error {
	valid := true
	for _, pattern := range yangType.Pattern {
		re, err := regexp.Compile("^" + pattern + "$")
		if err != nil {
			return errors.Wrap(err, "regexp.Compile")
		}
		if re.FindString(valueStr) != valueStr {
			valid = false
			break
		}
	}
	if !valid {
		return errors.Errorf("string value is not valid.")
	}
	return nil
}

func validateEnumValue(valueStr string, yangType *yang.YangType) error {
	valid := false
	for _, n := range yangType.Enum.Names() {
		if n == valueStr {
			valid = true
			break
		}
	}
	if !valid {
		return errors.Errorf(
			"enum value is not valid available=%+v",
			yangType.Enum.Names())
	}
	return nil
}

func validateNumberValue(valueStr string, yangType *yang.YangType) error {
	val := DBValue{Type: yangType.Kind}
	if err := val.SetFromString(valueStr); err != nil {
		return errors.Wrap(err, "SetFromString")
	}
	n, err := val.ToYangNumber()
	if err != nil {
		return errors.Wrap(err, "ToYangNumber")
	}
	for _, valRange := range yangType.Range {
		// Validate Min
		if n.Less(valRange.Min) {
			return errors.Errorf(
				"min validation failed min=%v input=%v",
				valRange.Min, n)
		}
		// Validate Min
		if valRange.Max.Less(*n) {
			return errors.Errorf(
				"max validation failed max=%v input=%v",
				valRange.Max, n)
		}
	}
	return nil
}

func resolveUnionTypes(yangTypes []*yang.YangType) []*yang.YangType {
	ret := []*yang.YangType{}
	for _, ytype := range yangTypes {
		switch ytype.Kind {
		case yang.Yunion:
			ret1 := resolveUnionTypes(ytype.Type)
			ret = append(ret, ret1...)
		default:
			ret = append(ret, ytype)
		}
	}
	return ret
}
