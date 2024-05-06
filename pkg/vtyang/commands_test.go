package vtyang

import (
	"reflect"
	"testing"

	"github.com/k0kubun/pp"
	"github.com/pkg/errors"
)

func TestInstallCompletionTree(t *testing.T) {
	root := &CompletionNode{
		Name: "",
		Childs: []*CompletionNode{
			{
				Name: "show",
				Childs: []*CompletionNode{
					{
						Name: "configuration",
						Childs: []*CompletionNode{
							newCR(),
						},
					},
				},
			},
		},
	}
	inputRoot := &CompletionNode{
		Childs: []*CompletionNode{
			{
				Name: "show",
				Childs: []*CompletionNode{
					{
						Name: "running-config",
						Childs: []*CompletionNode{
							newCR(),
						},
					},
				},
			},
		},
	}
	expect := &CompletionNode{
		Name: "",
		Childs: []*CompletionNode{
			{
				Name: "show",
				Childs: []*CompletionNode{
					{
						Name: "configuration",
						Childs: []*CompletionNode{
							newCR(),
						},
					},
					{
						Name: "running-config",
						Childs: []*CompletionNode{
							newCR(),
						},
					},
				},
			},
		},
	}
	if err := installCompletionTree(root, inputRoot); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(root, expect) {
		pp.Println("expect", expect)
		pp.Println("result", root)
		t.Errorf("missmatch")
	}
}

type TestDoCompletionTestCase struct {
	in  string
	out CompletionResult
}

func executeDoCompletionTestCase(testcase []TestDoCompletionTestCase, idx int) error {
	tc := testcase[idx]
	result := doCompletion(tc.in, len(tc.in))
	result.ResolvedXPath = nil
	for idx := range result.Items {
		result.Items[idx].Helper = ""
	}
	if !reflect.DeepEqual(result, tc.out) {
		pp.Println("expect", tc.out)
		pp.Println("result", result)
		return errors.Errorf("diff tc[%d] \"%s\"", idx, tc.in)
	}
	return nil
}

func TestDoCompletion01(t *testing.T) {
	// Init agent
	if err := InitAgent(AgentOpts{
		LogFile:     agentTestDefaultLogFile,
		RuntimePath: "/tmp/run/vtyang",
		YangPath:    []string{"./testdata/yang/basic"},
	}); err != nil {
		t.Fatal(err)
	}
	getCommandNodeCurrent().executeCommand("configure")

	// Testcase Decration
	testcases := []TestDoCompletionTestCase{
		{
			in: "set values",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "values"},
				},
			},
		},
		{
			in: "set values ",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "afi"},
					{Word: "bool"},
					{Word: "crypto"},
					{Word: "decimal"},
					{Word: "i08"},
					{Word: "i16"},
					{Word: "i32"},
					{Word: "i64"},
					{Word: "ipv4-address"},
					{Word: "ipv6-address"},
					{Word: "items"},
					{Word: "month"},
					{Word: "month-str"},
					{Word: "month-union"},
					{Word: "name"},
					{Word: "percentage"},
					{Word: "transport-proto"},
					{Word: "u08"},
					{Word: "u16"},
					{Word: "u32"},
					{Word: "u64"},
					{Word: "union-list"},
					{Word: "union-multiple"},
				},
			},
		},
		{
			in: "set values cry",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "crypto"},
				},
			},
		},
		{
			// TODO(slankdev): Okashii This behavior
			// out.InvalidArg must be true
			in: "set values cry ",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "VALUE"},
				},
			},
		},
		{
			in: "set values crypto",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "crypto"},
				},
			},
		},
		{
			in: "set values crypto ",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "main:aes"},
					{Word: "main:des3"},
				},
			},
		},
		{
			in: "set values crypto mai",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "main:aes"},
					{Word: "main:des3"},
				},
			},
		},
		{
			in: "set values crypto main:ae",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "main:aes"},
				},
			},
		},
		{
			in: "set values crypto main:aes",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "main:aes"},
				},
			},
		},
		{
			in: "set values crypto main:aes ",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "<cr>"},
				},
			},
		},
		{
			in: "set values items item4",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "item4"},
				},
			},
		},
		{
			in: "set values items item4 ",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "main:aes"},
					{Word: "main:des3"},
				},
			},
		},
		{
			in: "set values items item4 main:ae",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "main:aes"},
				},
			},
		},
		{
			in: "set values items item4 main:aes",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "main:aes"},
				},
			},
		},
		{
			in: "set values items item4 main:aes ",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "<cr>"},
					{Word: "algo2"},
					{Word: "description"},
				},
			},
		},
		{
			in: "set values items item5",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "item5"},
				},
			},
		},
		{
			in: "set values items item5 ",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "main:aes"},
					{Word: "main:des3"},
				},
			},
		},
		{
			in: "set values items item5 main:ae",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "main:aes"},
				},
			},
		},
		{
			in: "set values items item5 main:aes",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "main:aes"},
				},
			},
		},
		{
			in: "set values items item5 main:aes ",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "<cr>"},
					{Word: "main:aes"},
					{Word: "main:des3"},
				},
			},
		},
		{
			in: "set values items item5 main:aes main:de",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "main:des3"},
				},
			},
		},
		{
			in: "set values items item5 main:aes main:des3",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "main:des3"},
				},
			},
		},
		{
			in: "set values items item5 main:aes main:des3 ",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "<cr>"},
					{Word: "algo1"},
					{Word: "algo2"},
					{Word: "description"},
				},
			},
		},
		{
			in: "set values items item6",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "item6"},
				},
			},
		},
		{
			in: "set values items item6 ",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "main:aes"},
					{Word: "main:des3"},
				},
			},
		},
		{
			in: "set values items item6 main:ae",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "main:aes"},
				},
			},
		},
		{
			in: "set values items item6 main:aes",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "main:aes"},
				},
			},
		},
		{
			in: "set values items item6 main:aes ",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "<cr>"},
					{Word: "main:ftp"},
					{Word: "main:http"},
				},
			},
		},
		{
			in: "set values items item6 main:aes main:ht",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "main:http"},
				},
			},
		},
		{
			in: "set values items item6 main:aes main:http",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "main:http"},
				},
			},
		},
		{
			in: "set values items item6 main:aes main:http ",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "<cr>"},
					{Word: "algo"},
					{Word: "app"},
					{Word: "description"},
				},
			},
		},
		{
			in: "set values items item7 ",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "April"},
					{Word: "August"},
					{Word: "December"},
					{Word: "February"},
					{Word: "January"},
					{Word: "July"},
					{Word: "June"},
					{Word: "March"},
					{Word: "May"},
					{Word: "November"},
					{Word: "October"},
					{Word: "September"},
				},
			},
		},
		{
			in: "set values items item7 A",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "April"},
					{Word: "August"},
				},
			},
		},
		{
			in: "set values items item8",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "item8"},
				},
			},
		},
		{
			in: "set values items item8 ",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "NAME"},
				},
			},
		},
		{
			in: "set values items item8 hoge",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "NAME"},
				},
			},
		},
		{
			in: "set values items item8 hoge ",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "<cr>"},
					{Word: "description"},
				},
			},
		},
		{
			in: "set values items item9",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "item9"},
				},
			},
		},
		{
			in: "set values items item9 ",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "NAME"},
				},
			},
		},
		{
			in: "set values items item9 hoge",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "NAME"},
				},
			},
		},
		{
			in: "set values items item9 hoge ",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "<cr>"},
					{Word: "NAME"},
				},
			},
		},
		{
			in: "set values items item10 ",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "NAME"},
				},
			},
		},
		{
			in: "set values crypto",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "crypto"},
				},
			},
		},
		{
			in: "set values crypto ",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "main:aes"},
					{Word: "main:des3"},
				},
			},
		},
		{
			in: "set values crypto main:ae",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "main:aes"},
				},
			},
		},
		{
			in: "set values month-str",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "month-str"},
				},
			},
		},
		{
			in: "set values month-str ",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "April"},
					{Word: "August"},
					{Word: "December"},
					{Word: "February"},
					{Word: "January"},
					{Word: "July"},
					{Word: "June"},
					{Word: "March"},
					{Word: "May"},
					{Word: "November"},
					{Word: "October"},
					{Word: "September"},
				},
			},
		},
		{
			in: "set values month-str A",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "April"},
					{Word: "August"},
				},
			},
		},
		{
			in: "set values month-str April",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "April"},
				},
			},
		},
		{
			in: "set values month-str April ",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "<cr>"},
				},
			},
		},
		{
			in: "set values name",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "name"},
				},
			},
		},
		{
			in: "set values name ",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "VALUE"},
				},
			},
		},
		{
			in: "set values name hoge",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "VALUE"},
				},
			},
		},
		{
			in: "set values name hogefuga",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "VALUE"},
				},
			},
		},
		{
			in: "set values name hogefuga ",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "<cr>"},
				},
			},
		},
	}

	// Executes
	for idx := range testcases {
		t.Logf("execute tc[%d] \"%s\"", idx, testcases[idx].in)
		if err := executeDoCompletionTestCase(testcases, idx); err != nil {
			t.Errorf("%s\n", err)
		}
	}
}

func TestDoCompletion02(t *testing.T) {
	// Init agent
	if err := InitAgent(AgentOpts{
		LogFile:     agentTestDefaultLogFile,
		RuntimePath: "/tmp/run/vtyang",
		YangPath:    []string{"./testdata/yang/multi_module"},
	}); err != nil {
		t.Fatal(err)
	}
	getCommandNodeCurrent().executeCommand("configure")

	// Testcase Decration
	testcases := []TestDoCompletionTestCase{
		{
			in: "set system routing protocol ",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "main:connected"},
					{Word: "main:static"},
					{Word: "module1:bgp"},
					{Word: "module1:ospf"},
					{Word: "module2:isis"},
					{Word: "module2:ospf"},
				},
			},
		},
	}

	// Executes
	for idx := range testcases {
		t.Logf("execute tc[%d] \"%s\"", idx, testcases[idx].in)
		if err := executeDoCompletionTestCase(testcases, idx); err != nil {
			t.Errorf("fail tc[%d] err=\"%s\"\n", idx, err)
		}
	}
}

func TestDoCompletion03(t *testing.T) {
	// Init agent
	if err := InitAgent(AgentOpts{
		LogFile:     agentTestDefaultLogFile,
		RuntimePath: "/tmp/run/vtyang",
		YangPath:    []string{"./testdata/yang/frr_mgmtd_minimal"},
	}); err != nil {
		t.Fatal(err)
	}
	getCommandNodeCurrent().executeCommand("configure")

	// Testcase Decration
	testcases := []TestDoCompletionTestCase{
		{
			in: "set routing control-plane-protocols control-plane-protocol " +
				"frr-staticd:staticd staticd default staticd route-list ",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "NAME"},
				},
			},
		},
		{
			in: "set routing control-plane-protocols control-plane-protocol " +
				"frr-staticd:staticd staticd default staticd route-list 1.1.1.1/32",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "NAME"},
				},
			},
		},
		{
			in: "set routing control-plane-protocols control-plane-protocol " +
				"frr-staticd:staticd staticd default staticd route-list 1.1.1.1/32 ",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "<cr>"},
					{Word: "frr-routing:ipv4-flowspec"},
					{Word: "frr-routing:ipv4-labeled-unicast"},
					{Word: "frr-routing:ipv4-multicast"},
					{Word: "frr-routing:ipv4-unicast"},
					{Word: "frr-routing:ipv6-flowspec"},
					{Word: "frr-routing:ipv6-labeled-unicast"},
					{Word: "frr-routing:ipv6-multicast"},
					{Word: "frr-routing:ipv6-unicast"},
					{Word: "frr-routing:l2vpn-evpn"},
					{Word: "frr-routing:l2vpn-vpls"},
					{Word: "frr-routing:l3vpn-ipv4-multicast"},
					{Word: "frr-routing:l3vpn-ipv4-unicast"},
					{Word: "frr-routing:l3vpn-ipv6-multicast"},
					{Word: "frr-routing:l3vpn-ipv6-unicast"},
				},
			},
		},
		{
			in: "set routing control-plane-protocols control-plane-protocol " +
				"frr-staticd:staticd staticd default staticd route-list 1.1.1.1/32 frr",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "frr-routing:ipv4-flowspec"},
					{Word: "frr-routing:ipv4-labeled-unicast"},
					{Word: "frr-routing:ipv4-multicast"},
					{Word: "frr-routing:ipv4-unicast"},
					{Word: "frr-routing:ipv6-flowspec"},
					{Word: "frr-routing:ipv6-labeled-unicast"},
					{Word: "frr-routing:ipv6-multicast"},
					{Word: "frr-routing:ipv6-unicast"},
					{Word: "frr-routing:l2vpn-evpn"},
					{Word: "frr-routing:l2vpn-vpls"},
					{Word: "frr-routing:l3vpn-ipv4-multicast"},
					{Word: "frr-routing:l3vpn-ipv4-unicast"},
					{Word: "frr-routing:l3vpn-ipv6-multicast"},
					{Word: "frr-routing:l3vpn-ipv6-unicast"},
				},
			},
		},
		{
			in: "set routing control-plane-protocols control-plane-protocol " +
				"frr-staticd:staticd staticd default staticd route-list 1.1.1.1/32 " +
				"frr-routing:ipv4-unicast path-list 0 1 frr-nexthops nexthop ",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "blackhole"},
					{Word: "ifindex"},
					{Word: "ip4"},
					{Word: "ip4-ifindex"},
					{Word: "ip6"},
					{Word: "ip6-ifindex"},
				},
			},
		},
		{
			in: "set routing control-plane-protocols control-plane-protocol " +
				"frr-staticd:staticd staticd default staticd route-list 1.1.1.1/32 " +
				"frr-routing:ipv4-unicast path-list 0 1 frr-nexthops nexthop ip",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "ip4"},
					{Word: "ip4-ifindex"},
					{Word: "ip6"},
					{Word: "ip6-ifindex"},
				},
			},
		},
		{
			in: "set routing control-plane-protocols control-plane-protocol " +
				"frr-staticd:staticd staticd default staticd route-list 1.1.1.1/32 " +
				"frr-routing:ipv4-unicast path-list 0 1 frr-nexthops nexthop ip4",
			out: CompletionResult{
				Items: []CompletionItem{
					{Word: "ip4"},
					{Word: "ip4-ifindex"},
				},
			},
		},
	}

	// Executes
	for idx := range testcases {
		t.Logf("execute tc[%d] \"%s\"", idx, testcases[idx].in)
		if err := executeDoCompletionTestCase(testcases, idx); err != nil {
			t.Errorf("fail tc[%d] err=\"%s\"\n", idx, err)
		}
	}
}

func TestDoCompletion99(t *testing.T) {
	// Init agent
	if err := InitAgent(AgentOpts{
		LogFile:     agentTestDefaultLogFile,
		RuntimePath: "/tmp/run/vtyang",
		YangPath:    []string{"./testdata/yang/basic"},
	}); err != nil {
		t.Fatal(err)
	}
	getCommandNodeCurrent().executeCommand("configure")

	// Testcase Decration
	testcases := []TestDoCompletionTestCase{
		//
	}

	// Executes
	for idx := range testcases {
		t.Logf("execute tc[%d] \"%s\"", idx, testcases[idx].in)
		if err := executeDoCompletionTestCase(testcases, idx); err != nil {
			t.Errorf("fail tc[%d] err=\"%s\"\n", idx, err)
		}
	}
}
