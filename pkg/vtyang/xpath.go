package vtyang

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/openconfig/goyang/pkg/yang"
	"github.com/pkg/errors"

	"github.com/slankdev/vtyang/pkg/util"
)

type XWord struct {
	Module string
	Word   string
	// Keys
	Keys map[string]XWordKey
	// KeysIndex
	KeysIndex   []string `json:",omitempty"`
	Dbtype      DBNodeType
	Dbvaluetype yang.TypeKind
	Dbuniontype yang.TypeKind `json:"Dbuniontype,omitempty"`
	// UnionTypes
	UnionTypes []*yang.YangType `json:",omitempty"`
	// IdentityBase
	Identities []*yang.Identity `json:",omitempty"`
}

type XWordKey struct {
	Value DBValue
}

type XPath struct {
	Words []XWord
}

func (xp *XPath) TailIsLeaf() bool {
	if len(xp.Words) > 0 {
		return xp.Words[len(xp.Words)-1].Dbtype == Leaf
	}
	return false
}

func (xp *XPath) TailIsList() bool {
	if len(xp.Words) > 0 {
		return xp.Words[len(xp.Words)-1].Dbtype == List
	}
	return false
}

func (xp *XPath) Tail() *XWord {
	if len(xp.Words) > 0 {
		return &xp.Words[len(xp.Words)-1]
	}
	return nil
}

func (x XPath) String() string {
	s := ""
	for _, w := range x.Words {
		s = fmt.Sprintf("%s/%s:%s", s, w.Module, w.Word)
		if w.Keys != nil {
			for k, v := range w.Keys {
				s = fmt.Sprintf("%s[%s='%s']", s, k, v.Value.ToString())
			}
		}
	}
	return s
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

func validateIdentityrefValue(valueStr string, yangType *yang.YangType) error {
	possibleNames := []string{}
	for _, val := range yangType.IdentityBase.Values {
		possibleNames = append(possibleNames, val.Name)
	}
	if !util.StringInArray(valueStr, possibleNames) {
		return errors.Errorf("invalid identity possible=(%+v)", possibleNames)
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

type YangData struct {
	XPath XPath
	Value string
}

func CraftDBNode(datas []YangData) (*DBNode, error) {
	root, err := ReadFromJsonString("{}")
	if err != nil {
		return nil, errors.Wrap(err, "ReadFromJsonString")
	}
	dbm0 := NewDatabaseManager()
	dbm0.candidateRoot = root
	dbm0.candidateRoot.Type = Container
	for _, data := range datas {
		dbm0.SetNode(data.XPath, data.Value)
	}
	return dbm0.candidateRoot, nil
}

// TODO: dbm eliminate
func ParseXPathString(dbm *DatabaseManager, s string) (XPath, error) {
	var xpath XPath
	var err error
	for _, ent := range yangModuleDumpEntries() {
		module := &yang.Entry{}
		module.Dir = map[string]*yang.Entry{}
		module.Dir[ent.Name] = ent
		xpath, err = ParseXPathStringImpl(module, s)
		if err == nil {
			break
		}
	}
	if err != nil {
		return XPath{}, errors.Wrap(err, "ParseXPathStringImpl")
	}
	return xpath, nil
}

type dissectKV struct {
	k string
	v string
}

type dissectResult struct {
	module string
	word   string
	kvs    []dissectKV
}

func dissect(s string) (dissectResult, error) {
	re := regexp.MustCompile(`^([a-zA-Z0-9-]*:)?([a-zA-Z0-9-]*)(\[.*\])*$`)
	match := re.FindStringSubmatch(s)
	if len(match) != 4 {
		return dissectResult{}, errors.Errorf("invalid (%s) match-len=%d", s, len(match))
	}
	module := strings.Replace(match[1], ":", "", -1)
	word := match[2]
	keys := []dissectKV{}
	keysstr := match[3]
	re2 := regexp.MustCompile(`\[([a-zA-Z0-9-]*)='([a-zA-Z0-9-\./:]*)'\]`)
	match2 := re2.FindAllStringSubmatch(keysstr, -1)
	for _, match := range match2 {
		keys = append(keys, dissectKV{k: match[1], v: match[2]})
	}
	return dissectResult{module: module, word: word, kvs: keys}, nil
}

func xpathTokenize(s string) ([]dissectResult, error) {
	re := regexp.MustCompile(`[a-zA-Z0-9-:]+(\[[a-zA-Z0-9-:]+='[a-zA-Z0-9-\.:/]*'\])*`)
	matchs := re.FindAllStringSubmatch(s, -1)
	words := []string{}
	for _, m := range matchs {
		words = append(words, m[0])
	}
	results := []dissectResult{}
	for len(words) != 0 {
		result, err := dissect(words[0])
		if err != nil {
			return nil, errors.Wrap(err, "dissect")
		}
		results = append(results, result)
		words = words[1:]
	}
	return results, nil
}

func ParseXPathStringImpl(module *yang.Entry, s string) (XPath, error) {
	words, err := xpathTokenize(s)
	if err != nil {
		return XPath{}, errors.Wrap(err, "xpathTokenize")
	}
	xpath := XPath{}
	for len(words) != 0 {
		word := words[0].word
		keys := words[0].kvs
		xword := XWord{Word: word}
		var foundNode *yang.Entry = nil
		for n := range module.Dir {
			e := module.Dir[n]
			switch {
			case e.IsChoice():
				for _, ee := range e.Dir {
					switch {
					case ee.IsCase():
						for _, eee := range ee.Dir {
							if eee.Name == word {
								foundNode = eee
							}
						}
					default:
						panic("OKASHII")
					}
				}
			default:
				if n == word {
					foundNode = module.Dir[n]
				}
			}
			if foundNode != nil {
				break
			}
		}
		if foundNode == nil {
			return XPath{}, fmt.Errorf("entry %s is not found (%s)", word, s)
		}

		switch {
		case foundNode.IsContainer():
			xword.Dbtype = Container
		case foundNode.IsList():
			xword.Dbtype = List
			xword.Keys = map[string]XWordKey{}
			for _, w := range strings.Fields(foundNode.Key) {
				xword.KeysIndex = append(xword.KeysIndex, w)
				var keyLeafNode *yang.Entry
				for _, ee := range foundNode.Dir {
					if ee.Name == w {
						keyLeafNode = ee
						break
					}
				}
				if keyLeafNode == nil {
					return XPath{}, errors.Errorf("key(%s) not found", w)
				}

				// Parse list key-value
				valueStr := ""
				for _, key := range keys {
					if key.k == keyLeafNode.Name {
						valueStr = key.v
						break
					}
				}
				if valueStr == "" {
					return XPath{}, errors.Errorf("key(%s) value not found", w)
				}
				v := DBValue{Type: keyLeafNode.Type.Kind}
				if v.Type == yang.Yunion {
					unionType := yang.Ynone
					ytypes := resolveUnionTypes(keyLeafNode.Type.Type)
					for _, ytype := range ytypes {
						switch ytype.Kind {
						case yang.Ystring:
							if err := validateStringValue(valueStr, ytype); err == nil {
								unionType = ytype.Kind
							}
						case yang.Yenum:
							if err := validateEnumValue(valueStr, ytype); err == nil {
								unionType = ytype.Kind
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
							}
						default:
							panic(fmt.Sprintf("PANIC %s", ytype.Kind.String()))
						}
					}
					if unionType == yang.Ynone {
						panic(fmt.Sprintf("OKASHII valueStr=%s", valueStr))
					}
					v.UnionType = unionType
				}
				if err := v.SetFromString(valueStr); err != nil {
					return XPath{}, errors.Wrap(err, "SetFromstring")
				}
				xword.Keys[keyLeafNode.Name] = XWordKey{Value: v}
			}
		case foundNode.IsLeaf():
			xword.Dbtype = Leaf
			xword.Dbvaluetype = foundNode.Type.Kind
		}

		if foundNode.IsLeaf() && foundNode.Type.Kind == yang.Yunion {
			types := resolveUnionTypes(foundNode.Type.Type)
			xword.UnionTypes = types
		}

		mod, err := foundNode.InstantiatingModule()
		if err != nil {
			return XPath{}, errors.Wrap(err, "InstantiationgModule")
		}
		xword.Module = mod
		xpath.Words = append(xpath.Words, xword)
		words = words[1:]
		module = foundNode
	}

	return xpath, nil
}

func ParseXPathCli(dbm *DatabaseManager, args []string, tail []string,
	setmode bool) (XPath, []DBValue, []string, error) {
	if len(args) == 0 {
		return XPath{}, nil, tail, nil
	}
	xpath, val, err := ParseXPathArgs(dbm, args, setmode)
	if err != nil {
		tail = append(tail, args[len(args)-1])
		return ParseXPathCli(dbm, args[:len(args)-1], tail, setmode)
	}
	return xpath, val, tail, nil
}

func ParseXPathArgs(dbm *DatabaseManager, args []string,
	setmode bool) (XPath, []DBValue, error) {
	var xpath XPath
	var vals []DBValue
	var err error
	for _, ent := range yangModuleDumpEntries() {
		module := &yang.Entry{}
		module.Dir = map[string]*yang.Entry{}
		module.Dir[ent.Name] = ent
		xpath, vals, err = ParseXPathArgsImpl(module, args, setmode)
		if err == nil {
			break
		}
	}
	if err != nil {
		return XPath{}, nil, err
	}
	return xpath, vals, nil
}

func ParseXPathArgsImpl(module *yang.Entry, args []string,
	setmode bool) (XPath, []DBValue, error) {
	words := args
	xpath := XPath{}
	value := []DBValue{}
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
			return XPath{}, nil, errors.Errorf("entry %s is not found", words[0])
		}

		argumentCount := 1
		argumentExist := false
		switch {
		case foundNode.IsContainer():
			xword.Dbtype = Container
		case foundNode.IsLeaf():
			xword.Dbtype = Leaf
			xword.Dbvaluetype = foundNode.Type.Kind
			if setmode {
				if len(words) > 1 {
					v, err := validateValue(words[argumentCount], foundNode.Type)
					if err != nil {
						return XPath{}, nil, errors.Wrap(err, "validateValue")
					}
					value = append(value, v)
					argumentExist = true
				}
			}

		case foundNode.IsList():
			xword.Dbtype = List
			xword.Keys = map[string]XWordKey{}
			for _, w := range strings.Fields(foundNode.Key) {
				xword.KeysIndex = append(xword.KeysIndex, w)
				k := XWordKey{}
				xword.Keys[w] = k
			}
			if len(words) > 1 {
				for _, w := range strings.Fields(foundNode.Key) {
					if len(words[argumentCount:]) == 0 {
						break
					}
					var keyLeafNode *yang.Entry
					for _, ee := range foundNode.Dir {
						if ee.Name == w {
							keyLeafNode = ee
							break
						}
					}
					v, err := validateValue(words[argumentCount], keyLeafNode.Type)
					if err != nil {
						return XPath{}, nil, errors.Wrap(err, "validateValue(%s)")
					}
					tmp := xword.Keys[w]
					tmp.Value = v
					xword.Keys[w] = tmp
					argumentCount++
				}
				argumentCount--
				argumentExist = true
			}

		case foundNode.IsLeafList():
			if setmode {
				if len(words) < 2 {
					return XPath{}, nil, fmt.Errorf("is-leaf-list invalid args len")
				}
				for argumentCount < len(words) {
					v, err := validateValue(words[argumentCount], foundNode.Type)
					if err != nil {
						return XPath{}, nil, errors.Wrap(err, "validateValue(leaf-list)")
					}
					value = append(value, v)
					argumentCount++
				}
				argumentCount--
				argumentExist = true
			}
			xword.Dbtype = LeafList
			xword.Dbvaluetype = foundNode.Type.Kind
		default:
			panic("ASSERT")
		}

		if foundNode.IsLeaf() && foundNode.Type.Kind == yang.Yidentityref {
			xword.Identities = foundNode.Type.IdentityBase.Values
			for idx := range xword.Identities {
				xword.Identities[idx].Parent = nil
			}
		}
		if foundNode.IsLeaf() && foundNode.Type.Kind == yang.Yunion {
			types := resolveUnionTypes(foundNode.Type.Type)
			xword.UnionTypes = types
		}

		mod, err := foundNode.InstantiatingModule()
		if err != nil {
			return XPath{}, nil, errors.Wrap(err, "InstantiationgModule")
		}
		xword.Module = mod

		xpath.Words = append(xpath.Words, xword)
		words = words[1:]
		if argumentExist {
			words = words[argumentCount:]
		}
		module = foundNode
	}

	return xpath, value, nil
}

func validateValue(valueStr string, ytype *yang.YangType) (DBValue, error) {
	value := DBValue{}

	// Additional Validation for Union
	switch ytype.Kind {
	case yang.Yunion:
		validated := false
		ytypes := resolveUnionTypes(ytype.Type)
		for _, ytype := range ytypes {
			switch ytype.Kind {
			case yang.Ystring:
				if err := validateStringValue(valueStr, ytype); err == nil {
					validated = true
					value.Type = ytype.Kind
					if err := value.SetFromString(valueStr); err != nil {
						return DBValue{}, errors.Wrap(err, "SetFromString")
					}
				}
			case yang.Yenum:
				if err := validateEnumValue(valueStr, ytype); err == nil {
					validated = true
					value.Type = ytype.Kind
					if err := value.SetFromString(valueStr); err != nil {
						return DBValue{}, errors.Wrap(err, "SetFromString")
					}
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
					validated = true
					value.Type = ytype.Kind
					if err := value.SetFromString(valueStr); err != nil {
						return DBValue{}, errors.Wrap(err, "SetFromString")
					}
				}
			default:
				panic(fmt.Sprintf("PANIC %s", ytype.Kind.String()))
			}
			if validated {
				break
			}
		}
		if !validated {
			return DBValue{}, errors.Errorf("union not validated")
		}
	}

	// Additional Validation for String
	switch ytype.Kind {
	case yang.Ystring:
		if err := validateStringValue(valueStr, ytype); err != nil {
			return DBValue{}, errors.Wrap(err, "validateStringValue")
		}
	}

	// Additional Validation for Enum
	switch ytype.Kind {
	case yang.Yenum:
		if err := validateEnumValue(valueStr, ytype); err != nil {
			return DBValue{}, errors.Wrap(err, "validateEnumValue")
		}
	}

	// Additional Validatation for Number-types
	// - range (A)..(B)
	switch ytype.Kind {
	case yang.Yint8, yang.Yint16, yang.Yint32, yang.Yint64,
		yang.Yuint8, yang.Yuint16, yang.Yuint32, yang.Yuint64,
		yang.Ydecimal64:
		if err := validateNumberValue(valueStr, ytype); err != nil {
			return DBValue{}, errors.Wrap(err, "validateNumberValue")
		}
	}

	switch ytype.Kind {
	case yang.Yidentityref:
		if err := validateIdentityrefValue(valueStr, ytype); err != nil {
			return DBValue{}, errors.Wrap(err, "validateIdentityrefValue")
		}
	}

	if ytype.Kind != yang.Yunion {
		value.Type = ytype.Kind
		if err := value.SetFromString(valueStr); err != nil {
			return DBValue{}, errors.Wrap(err, "SetFromString")
		}
	}

	// Additional Validation for String-types
	// - length
	// - pattern
	// - pattern(invert-match)
	// TODO(slankdev): IMPLEMENT ME
	return value, nil
}
