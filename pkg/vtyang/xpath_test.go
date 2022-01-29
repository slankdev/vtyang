package vtyang

import (
	"reflect"
	"strings"
	"testing"

	"github.com/k0kubun/pp"
)

func init() {
	dbm = NewDatabaseManager()
	dbm.LoadYangModuleOrDie("./testdata")
	dbm.LoadDatabaseFromData(&DummyDBRoot)
}

func TestXPathParse(t *testing.T) {
	testcases := []struct {
		in    string
		xpath XPath
	}{
		{
			in: "/users/user['name'='eva']",
			xpath: XPath{
				words: []XWord{
					{
						dbtype: Container,
						word:   "users",
					},
					{
						dbtype: List,
						word:   "user",
						keys: map[string]string{
							"name": "eva",
						},
					},
				},
			},
		},
	}

	for _, tc := range testcases {
		xpath := NewXPathOrDie(tc.in)
		if !reflect.DeepEqual(xpath, tc.xpath) {
			pp.Println("in", tc.xpath)
			pp.Println("out", xpath)
			t.Errorf("missmatch deepequal in=%s", tc.in)
		}
		if xpath.String() != tc.in {
			t.Errorf("missmatch in=%s out=%s", tc.in, xpath.String())
		}
	}
}

func TestXPathParseCli(t *testing.T) {
	testcases := []struct {
		in    string
		val   string
		set   bool
		xpath XPath
	}{
		{
			in:  "users user eva age 200",
			val: "200",
			set: true,
			xpath: XPath{
				words: []XWord{
					{
						dbtype: Container,
						word:   "users",
					},
					{
						dbtype: List,
						word:   "user",
						keys: map[string]string{
							"name": "eva",
						},
					},
					{
						dbtype:      Leaf,
						word:        "age",
						dbvaluetype: YInteger,
					},
				},
			},
		},
	}

	for _, tc := range testcases {
		args := strings.Fields(tc.in)
		xpath, val, err := ParseXPathArgs(args, tc.set)
		ErrorOnDie(err)

		if !reflect.DeepEqual(xpath, tc.xpath) {
			pp.Println("in", tc.xpath)
			pp.Println("out", xpath)
			t.Errorf("missmatch deepequal in=%s", tc.in)
		}

		if val != tc.val {
			t.Errorf("missmatch val in=%s out=%s", tc.val, val)
		}
	}
}
