package vtyang

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/k0kubun/pp"
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
	dbm.db.root = testDummyDBRoot
	dbm.LoadYangModuleOrDie("./testdata")

	testcases := []struct {
		mod string
		in  string
		ptr *DBNode
	}{
		{
			mod: "account",
			in:  "/users/user['name'='alice']",
			ptr: &testDummyDBRoot.Childs[0].Childs[0].Childs[0],
		},
	}

	for _, tc := range testcases {
		xpath := NewXPathOrDie(mod, tc.in)
		node, err := dbm.GetNode(mod, xpath)
		ErrorOnDie(err)
		if node == nil {
			t.Errorf("not found")
		}

		if !reflect.DeepEqual(node, tc.ptr) {
			pp.Println(node)
			pp.Println(tc.ptr)
			t.Errorf("missmatch")
		}
	}
}

func TestDBNodeCreate(t *testing.T) {
	testcases := []struct {
		mod  string
		in   []string
		root DBNode
	}{
		{
			mod: "account",
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
		xpath := NewXPathOrDie(mod, tc.in[0])
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

func TestDBNodeMerge(t *testing.T) {
	testcases := []struct {
		layers []DBNode
		result DBNode
	}{
		{
			layers: []DBNode{
				{ // [0]
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
				}, // [0]
				{ // [1]
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
														String: "fuga",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}, // [1]
			},
			result: DBNode{
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
									{
										Type: Container,
										Childs: []DBNode{
											{
												Name: "name",
												Type: Leaf,
												Value: DBValue{
													Type:   YString,
													String: "fuga",
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
		result := DBNode{}
		for _, layer := range tc.layers {
			new, err := MergeDBNode(result, layer)
			ErrorOnDie(err)
			result = new
		}

		if !reflect.DeepEqual(result, tc.result) {
			fmt.Println(result.String())
			fmt.Println(tc.result.String())
			diff := cmp.Diff(result, tc.result)
			t.Errorf("Hogefunc differs: (-got +want)\n%s", diff)
		}
	}
}
