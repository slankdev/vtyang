package vtyang

import (
	"reflect"
	"strings"
	"testing"

	"github.com/k0kubun/pp"
)

var xpathTestDBRoot = DBNode{
	Type: Container,
	Childs: []DBNode{
		{
			Name: "users",
			Type: Container,
			Childs: []DBNode{
				{
					Name: "user",
					Type: List,
					Childs: []DBNode{
						{
							Type: Container,
							Childs: []DBNode{
								{
									Name: "name",
									Type: Leaf,
									Value: DBValue{
										Type:   YString,
										String: "hiroki",
									},
								},
								{
									Name: "age",
									Type: Leaf,
									Value: DBValue{
										Type:    YInteger,
										Integer: 26,
									},
								},
								{
									Name: "projects",
									Type: List,
									Childs: []DBNode{
										{
											Type: Container,
											Childs: []DBNode{
												{
													Name: "name",
													Type: Leaf,
													Value: DBValue{
														Type:   YString,
														String: "tennis",
													},
												},
												{
													Name: "finished",
													Type: Leaf,
													Value: DBValue{
														Type:    YBoolean,
														Boolean: true,
													},
												},
											},
										},
										{
											Type: Container,
											Childs: []DBNode{
												{
													Name: "name",
													Type: Leaf,
													Value: DBValue{
														Type:   YString,
														String: "driving",
													},
												},
												{
													Name: "finished",
													Type: Leaf,
													Value: DBValue{
														Type:    YBoolean,
														Boolean: false,
													},
												},
											},
										},
									},
								},
							},
						},
						{
							Type: Container,
							Childs: []DBNode{
								{
									Name: "name",
									Type: Leaf,
									Value: DBValue{
										Type:   YString,
										String: "slankdev",
									},
								},
								{
									Name: "age",
									Type: Leaf,
									Value: DBValue{
										Type:    YInteger,
										Integer: 36,
									},
								},
								{
									Name: "projects",
									Type: List,
									Childs: []DBNode{
										{
											Type: Container,
											Childs: []DBNode{
												{
													Name: "name",
													Type: Leaf,
													Value: DBValue{
														Type:   YString,
														String: "kloudnfv",
													},
												},
												{
													Name: "finished",
													Type: Leaf,
													Value: DBValue{
														Type:    YBoolean,
														Boolean: false,
													},
												},
											},
										},
										{
											Type: Container,
											Childs: []DBNode{
												{
													Name: "name",
													Type: Leaf,
													Value: DBValue{
														Type:   YString,
														String: "wide",
													},
												},
												{
													Name: "finished",
													Type: Leaf,
													Value: DBValue{
														Type:    YBoolean,
														Boolean: false,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	},
}

func TestXPathParse(t *testing.T) {
	dbm := NewDatabaseManager()
	dbm.LoadYangModuleOrDie("./testdata/yang/accounting")
	dbm.LoadDatabaseFromData(&xpathTestDBRoot)

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
						keys: map[string]DBValue{
							"name": {
								Type:   YString,
								String: "eva",
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range testcases {
		xpath := NewXPathOrDie(dbm, tc.in)
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
	dbm := NewDatabaseManager()
	dbm.LoadYangModuleOrDie("./testdata/yang/accounting")
	dbm.LoadDatabaseFromData(&xpathTestDBRoot)

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
						keys: map[string]DBValue{
							"name": {
								Type:   YString,
								String: "eva",
							},
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
		xpath, val, err := ParseXPathArgs(dbm, args, tc.set)
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
