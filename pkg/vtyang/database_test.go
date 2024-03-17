package vtyang

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/k0kubun/pp"
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
										Type:   YString,
										String: "alice",
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
										String: "bob",
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
	dbm.LoadYangModuleOrDie("./testdata")
	dbm.LoadDatabaseFromData(&testDummyDBRoot)

	testcases := []struct {
		in  string
		ptr *DBNode
	}{
		{
			in:  "/users/user['name'='alice']",
			ptr: &testDummyDBRoot.Childs[0].Childs[0].Childs[0],
		},
	}

	for _, tc := range testcases {
		xpath := NewXPathOrDie(dbm, tc.in)
		node, err := dbm.GetNode(xpath)
		ErrorOnDie(err)

		if !reflect.DeepEqual(node, tc.ptr) {
			pp.Println(node)
			pp.Println(tc.ptr)
			t.Errorf("missmatch")
		}
	}
}

func TestDBNodeCreate(t *testing.T) {
	dbm := NewDatabaseManager()
	dbm.LoadYangModuleOrDie("./testdata")
	dbm.LoadDatabaseFromData(&testDummyDBRoot)

	testcases := []struct {
		in   []string
		root DBNode
	}{
		{
			in: []string{
				"/users/user['name'='hoge']",
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
													Type:   YString,
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
		xpath := NewXPathOrDie(dbm, tc.in[0])
		root, err := xpath.CreateDBNodeTree()
		ErrorOnDie(err)

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
		ErrorOnDie(err)
		n, err := Interface2DBNode(m)
		ErrorOnDie(err)
		out := n.String()
		same, err := util.DeepEqualJSON(tc.in, out)
		ErrorOnDie(err)
		if !same {
			println(tc.in)
			println(out)
			t.Errorf("mismatch json")
		}
	}
}
