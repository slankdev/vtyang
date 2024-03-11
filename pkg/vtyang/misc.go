package vtyang

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/openconfig/goyang/pkg/yang"
	"github.com/slankdev/vtyang/pkg/util"
)

func YangTypeKind2YType(t yang.TypeKind) DBValueType {
	switch t {
	case yang.Yint32:
		return YInteger
	case yang.Yuint32:
		return YInteger
	case yang.Ystring:
		return YString
	case yang.Ybool:
		return YBoolean
	default:
		panic(fmt.Sprintf("TODO(%s)", t))
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
		fmt.Fprintf(stdout, "Error: %s\n", err.Error())
		return "{}"
	}
	var out bytes.Buffer
	if err = json.Indent(&out, b, "", "  "); err != nil {
		fmt.Fprintf(stdout, "Error: %s\n", err.Error())
		return "{}"
	}
	return out.String()
}

func jsonstring2map(s string) (interface{}, error) {
	m := map[string]interface{}{}
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return nil, err
	}
	return m, nil
}

func cat(args []string) string {
	s := ""
	for _, a := range args {
		s += fmt.Sprintf("%s ", a)
	}
	return s
}

func matchArgs(args []string, matchStr string) bool {
	matchArgs := strings.Fields(matchStr)
	if len(matchArgs) > len(args) {
		log.Printf("Unmatch %s v.s. %s\n", args, matchStr)
		return false
	}

	for i, _ := range matchArgs {
		if matchArgs[i] != args[i] {
			log.Printf("Unmatch %s v.s. %s\n", args, matchStr)
			return false
		}
	}

	log.Printf("Match %s v.s. %s\n", args, matchStr)
	return true
}

func newTable() *tablewriter.Table {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("  ")
	table.SetNoWhiteSpace(true)
	return table
}

func slicesContainsInt(s []int, v int) bool {
	for _, t := range s {
		if t == v {
			return true
		}
	}
	return false
}

func newCR() *CompletionNode {
	return &CompletionNode{
		Name: "<cr>",
	}
}

func setStdoutWithBuffer() *bytes.Buffer {
	buf := bytes.NewBufferString("")
	stdout = buf
	return buf
}
