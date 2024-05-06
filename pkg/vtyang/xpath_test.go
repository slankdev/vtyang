package vtyang

import (
	"reflect"
	"strings"
	"testing"

	"github.com/k0kubun/pp"
	"github.com/openconfig/goyang/pkg/yang"
	"github.com/slankdev/vtyang/pkg/util"
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
	var err error
	yangmodules, err = yangModulesPath([]string{"./testdata/yang/accounting"})
	if err != nil {
		t.Fatal(err)
	}

	dbm := NewDatabaseManager()
	dbm.LoadDatabaseFromData(&xpathTestDBRoot)

	testcases := []struct {
		in    string
		val   []DBValue
		set   bool
		xpath XPath
	}{
		{
			in: "users user eva age 200",
			val: []DBValue{
				{
					Type:  yang.Yint32,
					Int32: 200,
				},
			},
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
						KeysIndex: []string{
							"name",
						},
						Keys: map[string]XWordKey{
							"name": {
								ytype: yang.YangType{
									Name: "string",
									Kind: yang.Ystring,
									Base: nil,
								},
								Value: DBValue{
									Type:   yang.Ystring,
									String: "eva",
								},
							},
						},
					},
					{
						Module:      "account",
						Dbtype:      Leaf,
						Word:        "age",
						Dbvaluetype: yang.Yint32,
						ytype: yang.YangType{
							Name: yang.Yint32.String(),
							Kind: yang.Yint32,
							Range: yang.YangRange{
								yang.YRange{
									Min: yang.Number{
										Value:          0x0000000080000000,
										FractionDigits: 0x00,
										Negative:       true,
									},
									Max: yang.Number{
										Value:          0x000000007fffffff,
										FractionDigits: 0x00,
										Negative:       false,
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
		args := strings.Fields(tc.in)
		xpath, val, err := ParseXPathArgs(dbm, args, tc.set)
		util.PanicOnErr(err)

		// Post process
		for idx := range xpath.Words {
			if xpath.Words[idx].Dbtype == List {
				for idx2 := range xpath.Words[idx].Keys {
					tmp := xpath.Words[idx].Keys[idx2]
					tmp.ytype.Base = nil
					tmp.ytype.Root = nil
					xpath.Words[idx].Keys[idx2] = tmp
				}
			}
			if xpath.Words[idx].Dbtype == Leaf {
				xpath.Words[idx].ytype.Base = nil
				xpath.Words[idx].ytype.Root = nil
			}
		}

		// Compare
		if !reflect.DeepEqual(xpath, tc.xpath) {
			pp.Println("xpath-expect", tc.xpath)
			pp.Println("xpath-result", xpath)
			t.Errorf("missmatch deepequal in=%s", tc.in)
		}

		if len(val) != len(tc.val) {
			t.Errorf("missmatch val-len in=%d out=%d", len(tc.val), len(val))
		}
		if val[0] != tc.val[0] {
			t.Errorf("missmatch val in=%s out=%s", tc.val[0].ToString(), val[0].ToString())
		}
	}
}

func Test_xpathTokenize(t *testing.T) {
	testcases := []struct {
		in  string
		out []dissectResult
	}{
		{
			in: "/frr-routing:routing/control-plane-protocols/" +
				"control-plane-protocol[type='frr-staticd:staticd']" +
				"[name='staticd'][vrf='default']/" +
				"frr-staticd:staticd/route-list[prefix='1.1.1.1/32']" +
				"[afi-safi='frr-routing:ipv4-unicast']/prefix",
			out: []dissectResult{
				{
					module: "frr-routing",
					word:   "routing",
					kvs:    []dissectKV{},
				},
				{
					module: "",
					word:   "control-plane-protocols",
					kvs:    []dissectKV{},
				},
				{
					module: "",
					word:   "control-plane-protocol",
					kvs: []dissectKV{
						{k: "type", v: "frr-staticd:staticd"},
						{k: "name", v: "staticd"},
						{k: "vrf", v: "default"},
					},
				},
				{
					module: "frr-staticd",
					word:   "staticd",
					kvs:    []dissectKV{},
				},
				{
					module: "",
					word:   "route-list",
					kvs: []dissectKV{
						{k: "prefix", v: "1.1.1.1/32"},
						{k: "afi-safi", v: "frr-routing:ipv4-unicast"},
					},
				},
				{
					module: "",
					word:   "prefix",
					kvs:    []dissectKV{},
				},
			},
		},
	}

	for idx, tc := range testcases {
		results, err := xpathTokenize(tc.in)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(results, tc.out) {
			t.Errorf("tc[%d] DIFF", idx)
		}
	}
}

func Test_dissect(t *testing.T) {
	type kv struct {
		k string
		v string
	}

	testcases := []struct {
		in  string
		out struct {
			module string
			word   string
			keys   []kv
		}
	}{
		{
			in: "control-plane-protocol[type='frr-staticd:staticd'][name='staticd'][vrf='default']",
			out: struct {
				module string
				word   string
				keys   []kv
			}{
				module: "",
				word:   "control-plane-protocol",
				keys: []kv{
					{k: "type", v: "frr-staticd:staticd"},
					{k: "name", v: "staticd"},
					{k: "vrf", v: "default"},
				},
			},
		},
		{
			in: "hoge[prefix='1.1.1.1/32']",
			out: struct {
				module string
				word   string
				keys   []kv
			}{
				module: "",
				word:   "hoge",
				keys: []kv{
					{k: "prefix", v: "1.1.1.1/32"},
				},
			},
		},
		{
			in: "module:control-plane-protocol[type='frr-staticd:staticd'][name='staticd'][vrf='default']",
			out: struct {
				module string
				word   string
				keys   []kv
			}{
				module: "module",
				word:   "control-plane-protocol",
				keys: []kv{
					{k: "type", v: "frr-staticd:staticd"},
					{k: "name", v: "staticd"},
					{k: "vrf", v: "default"},
				},
			},
		},
		{
			in: "frr-routing:routing",
			out: struct {
				module string
				word   string
				keys   []kv
			}{
				module: "frr-routing",
				word:   "routing",
				keys:   []kv{},
			},
		},
	}

	for idx, tc := range testcases {
		result, err := dissect(tc.in)
		if err != nil {
			t.Fatal(err)
		}
		module := result.module
		word := result.word
		keys := result.kvs
		if word != tc.out.word {
			t.Errorf("tc[%d] unexpected word expect=%s result=%s", idx, tc.out.word, word)
		}
		if module != tc.out.module {
			t.Errorf("tc[%d] unexpected module expect=%s result=%s", idx, tc.out.module, module)
		}
		if len(keys) != len(tc.out.keys) {
			t.Errorf("tc[%d] unexpected len(keys) expect=%d result=%d",
				idx, len(tc.out.keys), len(keys))
		}
		for i := 0; i < len(keys); i++ {
			if keys[i].k != tc.out.keys[i].k {
				t.Errorf("tc[%d] unexpected keys[%d].k expect=%s result=%s", idx, i,
					tc.out.keys[i].k, keys[i].k)
			}
			if keys[i].v != tc.out.keys[i].v {
				t.Errorf("tc[%d] unexpected keys[%d].v expect=%s result=%s", idx, i,
					tc.out.keys[i].v, keys[i].v)
			}
		}
	}
}
