package vtyang

import (
	"reflect"
	"strings"
	"testing"

	"github.com/k0kubun/pp"
	"github.com/openconfig/goyang/pkg/yang"
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
										Type:   yang.Ystring,
										String: "hiroki",
									},
								},
								{
									Name: "age",
									Type: Leaf,
									Value: DBValue{
										Type:  yang.Yint32,
										Int32: 26,
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
														Type:   yang.Ystring,
														String: "tennis",
													},
												},
												{
													Name: "finished",
													Type: Leaf,
													Value: DBValue{
														Type:    yang.Ybool,
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
														Type:   yang.Ystring,
														String: "driving",
													},
												},
												{
													Name: "finished",
													Type: Leaf,
													Value: DBValue{
														Type:    yang.Ybool,
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
										Type:   yang.Ystring,
										String: "slankdev",
									},
								},
								{
									Name: "age",
									Type: Leaf,
									Value: DBValue{
										Type:  yang.Yint32,
										Int32: 36,
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
														Type:   yang.Ystring,
														String: "kloudnfv",
													},
												},
												{
													Name: "finished",
													Type: Leaf,
													Value: DBValue{
														Type:    yang.Ybool,
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
														Type:   yang.Ystring,
														String: "wide",
													},
												},
												{
													Name: "finished",
													Type: Leaf,
													Value: DBValue{
														Type:    yang.Ybool,
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
				Words: []XWord{
					{
						Module: "account",
						Dbtype: Container,
						Word:   "users",
					},
					{
						Module: "account",
						Dbtype: List,
						Word:   "user",
						Keys: map[string]DBValue{
							"name": {
								Type:   yang.Ystring,
								String: "eva",
							},
						},
					},
					{
						Module:      "account",
						Dbtype:      Leaf,
						Word:        "age",
						Dbvaluetype: yang.Yint32,
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
