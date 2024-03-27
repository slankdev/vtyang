package main

// XXX(slankdev): ygot generator isn't working fine....
////go:generate generator -package_name=yang -path=yang -output_file=../../pkg/linux-agent/yang/generated.go -generate_fakeroot -fakeroot_name=device -shorten_enum_leaf_names -typedef_enum_with_defmod ../../yang/linux-agent.yang

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/k0kubun/pp"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	vtyangapi "github.com/slankdev/vtyang/pkg/grpc/api"
	"github.com/slankdev/vtyang/pkg/linux-agent/yang"
	"github.com/slankdev/vtyang/pkg/util"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	if err := NewCommand().Execute(); err != nil {
		os.Exit(1)
	}
}

var (
	clioptNetns   string
	clioptConnect string
)

func NewCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:  "linux-agent",
		RunE: f,
	}
	rootCmd.Flags().StringVar(&clioptNetns, "netns", "ns0",
		"name of network namespace")
	rootCmd.Flags().StringVar(&clioptConnect, "connect",
		"192.168.64.1:8080",
		"vtyang server")
	rootCmd.AddCommand(util.NewCommandCompletion(rootCmd))
	rootCmd.AddCommand(util.NewCommandVersion())
	return rootCmd
}

func f(cmd *cobra.Command, args []string) error {
	conn, err := grpc.Dial(
		clioptConnect,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return err
	}
	pp.Println("connected")
	defer conn.Close()
	client := vtyangapi.NewGreetingServiceClient(conn)
	stream, err := client.HelloStream(context.Background())
	if err != nil {
		return err
	}
	loop := true
	for loop {
		res, err := stream.Recv()
		if err != nil {
			fmt.Println(err.Error())
			loop = false
			continue
		}
		device := yang.Device{}
		if err := json.Unmarshal([]byte(res.Data), &device); err != nil {
			return err
		}
		pp.Println(device)
		if err := validate(&device); err != nil {
			return err
		}
		if err := commit(&device); err != nil {
			return err
		}
	}
	return nil
}

func validate(_ *yang.Device) error {
	// TODO(slankdev): it's not implemented
	return nil
}

func commit(device *yang.Device) error {
	foundIfaces := []string{}
	for _, iface := range device.Interfaces.Interface {
		if iface.Name == nil {
			continue
		}
		pp.Println(*iface.Name)
		foundIfaces = append(foundIfaces, *iface.Name)

		// get current status
		addrs := []Addr{}
		if err := util.LocalExecutefJsonMarshal(
			&addrs, "ip -j -n %s addr show dev %s",
			clioptNetns, *(iface.Name)); err != nil {
			return err
		}
		if len(addrs) != 1 {
			return fmt.Errorf("undefined case")
		}

		// ip addr add
		found := false
		pp.Println("DEBUG1", addrs)
		for _, addrInfo := range addrs[0].AddrInfo {
			addr1 := *iface.Address
			addr2 := fmt.Sprintf("%s/%d", addrInfo.Local, addrInfo.Prefixlen)
			pp.Println("a1: ", addr1)
			pp.Println("a2: ", addr2)
			if addr1 == addr2 {
				found = true
				break
			}
		}
		if !found {
			if _, err := util.LocalExecutef("ip netns exec %s ip addr add %s dev %s",
				clioptNetns, *iface.Address, *iface.Name); err != nil {
				return err
			}
		}

		// ip addr del
		for _, addrInfo := range addrs[0].AddrInfo {
			if addrInfo.Family == "inet" {
				keep := false
				if fmt.Sprintf("%s/%d", addrInfo.Local, addrInfo.Prefixlen) == *iface.Address {
					keep = true
				}
				if !keep {
					if _, err := util.LocalExecutef(
						"ip netns exec %s ip addr del %s/%d dev %s",
						clioptNetns, addrInfo.Local, addrInfo.Prefixlen,
						*iface.Name); err != nil {
						return err
					}
				}
			}
		}
	}

	// ip addr flush
	// ip link set down
	allAddrs := []Addr{}
	if err := util.LocalExecutefJsonMarshal(
		&allAddrs, " sudo ip -j -n %s addr",
		clioptNetns); err != nil {
		return err
	}
	for _, addr := range allAddrs {
		found := false
		for _, iface := range foundIfaces {
			if iface == addr.Ifname {
				found = true
				break
			}
		}
		if !found {
			if _, err := util.LocalExecutef(
				"ip netns exec %s ip addr flush dev %s",
				clioptNetns, addr.Ifname); err != nil {
				return err
			}
			if _, err := util.LocalExecutef(
				"ip netns exec %s ip link set down dev %s",
				clioptNetns, addr.Ifname); err != nil {
				return err
			}
		}
	}

	return nil
}

// sudo ip -j -n ns0 addr show dev dum0 | jq
// [
//
//	{
//	  "ifindex": 2,
//	  "ifname": "dum0",
//	  "mtu": 1500,
//	  "operstate": "DOWN",
//	  "link_type": "ether",
//	  "address": "d6:3e:c5:91:e9:0c",
//	  "addr_info": [
//	    {
//	      "family": "inet",
//	      "local": "10.0.0.1",
//	      "prefixlen": 24
//	    }
//	  ]
//	}
//
// ]
type Addr struct {
	Ifindex  int    `json:"ifindex"`
	Ifname   string `json:"ifname"`
	MTU      int    `json:"mtu"`
	Address  string `json:"address"`
	AddrInfo []struct {
		Family    string `json:"family"`
		Local     string `json:"local"`
		Prefixlen int    `json:"prefixlen"`
	} `json:"addr_info"`
}
