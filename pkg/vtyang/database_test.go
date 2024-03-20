package vtyang

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/k0kubun/pp"
	"github.com/nsf/jsondiff"
	"github.com/openconfig/goyang/pkg/yang"

	"github.com/slankdev/vtyang/pkg/util"
)

var testDummyDBRoot = DBNode{
	Name: "<root>",
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
										String: "alice",
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
										String: "bob",
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
							},
						},
					},
				},
			},
		},
	},
}

func TestDBNodeGet(t *testing.T) {
	dbm := NewDatabaseManager()
	dbm.LoadYangModuleOrDie("./testdata/yang/accounting")
	dbm.LoadDatabaseFromData(&testDummyDBRoot)

	testcases := []struct {
		in  string
		ptr *DBNode
	}{
		{
			in:  "/users/user[name='alice']",
			ptr: &testDummyDBRoot.Childs[0].Childs[0].Childs[0],
		},
	}

	for _, tc := range testcases {
		xpath, err := ParseXPathString(dbm, tc.in)
		util.PanicOnErr(err)
		node, err := dbm.GetNode(xpath)
		util.PanicOnErr(err)

		if !reflect.DeepEqual(node, tc.ptr) {
			pp.Println(node)
			pp.Println(tc.ptr)
			t.Errorf("missmatch")
		}
	}
}

func TestDBNodeCreate(t *testing.T) {
	dbm := NewDatabaseManager()
	dbm.LoadYangModuleOrDie("./testdata/yang/accounting")
	dbm.LoadDatabaseFromData(&testDummyDBRoot)

	testcases := []struct {
		in   []string
		root DBNode
	}{
		{
			in: []string{
				"/users/user[name='hoge']",
			},
			root: DBNode{
				Name: "<root>",
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
													String: "hoge",
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

	for _, tc := range testcases {
		xpath, err := ParseXPathString(dbm, tc.in[0])
		util.PanicOnErr(err)
		root, err := xpath.CreateDBNodeTree()
		util.PanicOnErr(err)

		if !reflect.DeepEqual(root, &tc.root) {
			fmt.Println(root.String())
			fmt.Println(tc.root.String())
			diff := cmp.Diff(root, tc.root)
			t.Errorf("Hogefunc differs: (-got +want)\n%s", diff)
		}
	}
}

func TestDBNodeJson(t *testing.T) {
	testcases := []struct {
		in string
	}{
		{
			in: `
{
	"users": {
		"user": [
			{
				"age": 26,
				"name": "hiroki",
				"projects": [
					{
						"finished": true,
						"name": "tennis"
					}
				]
			}
		]
	}
}
`,
		},
		{
			in: `{"users": {"user": [{"name": "hiroki"}]}}`,
		},

		// TODO(slankdev): bellow's data will be crash...
		// {
		// 	in: `{"users": {}}`,
		// },
	}

	for _, tc := range testcases {
		m := map[string]interface{}{}
		err := json.Unmarshal([]byte(tc.in), &m)
		util.PanicOnErr(err)
		n, err := Interface2DBNode(m)
		util.PanicOnErr(err)
		out := n.String()
		same, err := util.DeepEqualJSON(tc.in, out)
		util.PanicOnErr(err)
		if !same {
			println(tc.in)
			println(out)
			t.Errorf("mismatch json")
		}
	}
}

func TestSetNode(t *testing.T) {
	db := &DBNode{
		Type: "container",
		Childs: []DBNode{
			{
				Name: "isis",
				Type: "container",
				Childs: []DBNode{
					{
						Name: "instance",
						Type: "list",
						Childs: []DBNode{
							{
								Type: "container",
								Childs: []DBNode{
									{
										Name: "area-tag",
										Type: "leaf",
										Value: DBValue{
											Type:   yang.Ystring,
											String: "1",
										},
									},
									{
										Name: "vrf",
										Type: "leaf",
										Value: DBValue{
											Type:   yang.Ystring,
											String: "default",
										},
									},
									{
										Name: "description",
										Type: "leaf",
										Value: DBValue{
											Type:   yang.Ystring,
											String: "area1-default-hoge",
										},
									},
								},
							},
							{
								Type: "container",
								Childs: []DBNode{
									{
										Name: "area-tag",
										Type: "leaf",
										Value: DBValue{
											Type:   yang.Ystring,
											String: "1",
										},
									},
									{
										Name: "vrf",
										Type: "leaf",
										Value: DBValue{
											Type:   yang.Ystring,
											String: "vrf0",
										},
									},
									{
										Name: "description",
										Type: "leaf",
										Value: DBValue{
											Type:   yang.Ystring,
											String: "area1-vrf0-hoge",
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

	xpath := XPath{
		Words: []XWord{
			{
				Word:   "isis",
				Dbtype: "container",
			},
			{
				Word: "instance",
				Keys: map[string]DBValue{
					"area-tag": {
						Type:   yang.Ystring,
						String: "1",
					},
					"vrf": {
						Type:   yang.Ystring,
						String: "default",
					},
				},
				Dbtype: "list",
			},
			{
				Word:        "description",
				Keys:        map[string]DBValue{},
				Dbtype:      "leaf",
				Dbvaluetype: yang.Ystring,
			},
		},
	}

	_, _ = db, xpath
}

func TestDBNodeDeepCopy(t *testing.T) {
	original := &DBNode{
		Name: "",
		Type: Container,
		Childs: []DBNode{
			{
				Name: "child1",
				Type: Leaf,
				Value: DBValue{
					Type:   yang.Ystring,
					String: "value1",
				},
			},
			{
				Name: "child2",
				Type: Leaf,
				Value: DBValue{
					Type:  yang.Yint32,
					Int32: 42,
				},
			},
		},
	}

	copy := original.DeepCopy()
	if !reflect.DeepEqual(original, copy) {
		t.Errorf("DeepCopy failed: original and copy are not equal")
	}

	// Modify the copy and make sure it doesn't affect the original
	copy.Name = "modified"
	copy.Childs[0].Value.String = "modified value"
	if reflect.DeepEqual(original, copy) {
		t.Errorf("DeepCopy failed: modifying the copy modified the original")
	}
}

func Test_DBValue_ToAbsoluteNumber(t *testing.T) {
	inputs := []struct {
		Abs     uint64
		DBValue DBValue
	}{
		{
			Abs: 128,
			DBValue: DBValue{
				Type: yang.Yint8,
				Int8: -128,
			},
		},
		{
			Abs: 127,
			DBValue: DBValue{
				Type: yang.Yint8,
				Int8: 127,
			},
		},
		{
			Abs: 0,
			DBValue: DBValue{
				Type:  yang.Yuint8,
				Uint8: 0,
			},
		},
		{
			Abs: 255,
			DBValue: DBValue{
				Type:  yang.Yuint8,
				Uint8: 255,
			},
		},
	}

	for _, input := range inputs {
		v, _, _ := input.DBValue.ToAbsoluteNumber()
		if v != uint64(input.Abs) {
			t.Errorf("OKASHII %v v.s. %v", v, input.Abs)
		}
	}
}

func Test_SetNode(t *testing.T) {
	xpath := XPath{
		Words: []XWord{
			{
				Module: "frr-interface",
				Word:   "lib",
				Dbtype: Container,
			},
			{
				Module: "frr-interface",
				Word:   "interface",
				Dbtype: List,
				Keys: map[string]DBValue{
					"name": {
						Type:   yang.Ystring,
						String: "dum0",
					},
				},
			},
			{
				Module:      "frr-interface",
				Word:        "description",
				Dbtype:      Leaf,
				Dbvaluetype: yang.Ystring,
			},
		},
	}
	value := "hoge"
	out, err := CraftDBNode([]YangData{
		{
			XPath: xpath,
			Value: value,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	const expect = `{
  "lib": {
    "interface": [
      {
        "description": "hoge",
        "name": "dum0"
      }
    ]
  }
}`

	// Compare as json string
	opts := jsondiff.DefaultConsoleOptions()
	opts.Indent = "  "
	opt, diff := jsondiff.Compare([]byte(expect), []byte(out.String()), &opts)
	if opt != jsondiff.FullMatch {
		fmt.Printf("exp: \n%s\n", expect)
		fmt.Printf("out: \n%s\n", out)
		fmt.Printf("diff: \n%s\n", diff)
		t.Fatal("unexpected output")
	}
}
