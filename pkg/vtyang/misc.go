package vtyang

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/openconfig/goyang/pkg/yang"
	"github.com/slankdev/vtyang/pkg/util"
)

func YangTypeKind2YType(t yang.TypeKind) DBValueType {
	switch t {
	case yang.Yint32:
		return YInteger
	case yang.Ystring:
		return YString
	case yang.Ybool:
		return YBoolean
	default:
		panic("TODO")
	}
}

func name(s string) string {
	ret := util.SplitMultiSep(s, []string{"'", "[", "]", "="})
	return ret[0]
}

func hasKV(s string) bool {
	ret := util.SplitMultiSep(s, []string{"'", "[", "]", "="})
	return len(ret) == 3
}

func key(s string) string {
	ret := util.SplitMultiSep(s, []string{"'", "[", "]", "="})
	if len(ret) != 3 {
		panic(s)
	}
	return ret[1]
}

func val(s string) string {
	ret := util.SplitMultiSep(s, []string{"'", "[", "]", "="})
	if len(ret) != 3 {
		panic(s)
	}
	return ret[2]
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
