package vtyang

import (
	"encoding/json"
	"fmt"
	"os"
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
											Type:   "string",
											String: "1",
										},
									},
									{
										Name: "vrf",
										Type: "leaf",
										Value: DBValue{
											Type:   "string",
											String: "default",
										},
									},
									{
										Name: "description",
										Type: "leaf",
										Value: DBValue{
											Type:   "string",
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
											Type:   "string",
											String: "1",
										},
									},
									{
										Name: "vrf",
										Type: "leaf",
										Value: DBValue{
											Type:   "string",
											String: "vrf0",
										},
									},
									{
										Name: "description",
										Type: "leaf",
										Value: DBValue{
											Type:   "string",
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
		words: []XWord{
			{
				word:   "isis",
				dbtype: "container",
			},
			{
				word: "instance",
				keys: map[string]string{
					"area-tag": "1",
					"vrf":      "default",
				},
				dbtype: "list",
			},
			{
				word:        "description",
				keys:        map[string]string{},
				dbtype:      "leaf",
				dbvaluetype: "string",
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
					Type:   YString,
					String: "value1",
				},
			},
			{
				Name: "child2",
				Type: Leaf,
				Value: DBValue{
					Type:    YInteger,
					Integer: 42,
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

func TestFilterDbWithModule(t *testing.T) {
	input := &DBNode{
		Name: "",
		Type: Container,
		Childs: []DBNode{
			{
				Name: "bgp",
				Type: Container,
				Childs: []DBNode{
					{
						Name: "as-number",
						Type: Leaf,
						Value: DBValue{
							Type:    YInteger,
							Integer: 65001,
						},
					},
				},
			},
			{
				Name: "isis",
				Type: Container,
				Childs: []DBNode{
					{
						Name: "ignored",
						Type: Leaf,
						Value: DBValue{
							Type:    YInteger,
							Integer: 65001,
						},
					},
					{
						Name: "instance",
						Type: List,
						Childs: []DBNode{
							{
								Name: "",
								Type: Container,
								Childs: []DBNode{
									{
										Name: "area-tag",
										Type: Leaf,
										Value: DBValue{
											Type:   YString,
											String: "1",
										},
									},
									{
										Name: "vrf",
										Type: Leaf,
										Value: DBValue{
											Type:   YString,
											String: "default",
										},
									},
									{
										Name: "area-address",
										Type: LeafList,
										Value: DBValue{
											Type: YStringArray,
											StringArray: []string{
												"10.0000.0000.0000.0000.0000.0000.0000.0000.0000.00",
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

	expected := &DBNode{
		Name: "",
		Type: Container,
		Childs: []DBNode{
			{
				Name: "frr-isisd:isis",
				Type: Container,
				Childs: []DBNode{
					{
						Name: "instance",
						Type: List,
						Childs: []DBNode{
							{
								Name: "",
								Type: Container,
								Childs: []DBNode{
									{
										Name: "area-tag",
										Type: Leaf,
										Value: DBValue{
											Type:   YString,
											String: "1",
										},
									},
									{
										Name: "vrf",
										Type: Leaf,
										Value: DBValue{
											Type:   YString,
											String: "default",
										},
									},
									{
										Name: "area-address",
										Type: LeafList,
										Value: DBValue{
											Type: YStringArray,
											StringArray: []string{
												"10.0000.0000.0000.0000.0000.0000.0000.0000.0000.00",
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

	// Preparation
	GlobalOptRunFilePath = RUNTIME_PATH
	if util.FileExists(getDatabasePath()) {
		if err := os.Remove(getDatabasePath()); err != nil {
			t.Error(err)
		}
	}

	// Initializing Agent
	if err := InitAgent(RUNTIME_PATH,
		"../../yang.frr/"); err != nil {
		t.Fatal(err)
	}

	result, err := filterDbWithModule(input, "frr-isisd")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(result.String())

	diff := DBNodeDiff(result, expected)
	if diff != "" {
		t.Fatalf("diff %s\n", diff)
	}
}

func TestLoadDatabaseFromFile(t *testing.T) {
	testcases := []TestCaseForTestAgent{
		{
			Inputs: []string{
				"show running-config-frr",
			},
			Output: `{
        "frr-isisd:isis": {
          "instance": [
            {
              "area-address": [
                "10.0000.0000.0000.0000.0000.0000.0000.0000.0000.00"
              ],
              "area-tag": "1",
              "vrf": "default"
            }
          ]
        }
			}`,
		},
	}

	// Initializing Agent
	if err := InitAgent(
		"./testdata/runtime1",
		"../../yang.frr/"); err != nil {
		t.Fatal(err)
	}

	for idx, tc := range testcases {
		buf := setStdoutWithBuffer()
		for _, input := range tc.Inputs {
			t.Logf("Testcase[%d] executing %s", idx, input)
			getCommandNodeCurrent().executeCommand(input)
		}
		result := buf.String()
		eq, err := util.DeepEqualJSON(result, tc.Output)
		if err != nil {
			t.Fatal(err)
		}
		if !eq {
			t.Errorf("Unexpected output")
			for _, input := range tc.Inputs {
				t.Errorf("input %+v", input)
			}
			t.Errorf("expect(len=%d) %+v", len(tc.Output), tc.Output)
			t.Errorf("result(len=%d) %+v", len(result), result)
			t.Fatal("quiting test with FAILED result")
		}
		t.Logf("Testcase[%d] output check is succeeded", idx)
	}
}
