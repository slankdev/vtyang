# Cheatsheet

## yang.TypeKind iota

https://pkg.go.dev/github.com/openconfig/goyang/pkg/yang#pkg-constants

```
0x00	00 Ynone = TypeKind(iota)
0x01	01 Yint8
0x02	02 Yint16
0x03	03 Yint32
0x04	04 Yint64
0x05	05 Yuint8
0x06	06 Yuint16
0x07	07 Yuint32
0x08	08 Yuint64
0x09	09 Ybinary
0x0a	10 Ybits
0x0b	11 Ybool
0x0c	12 Ydecimal64
0x0d	13 Yempty
0x0e	14 Yenum
0x0f	15 Yidentityref
0x10	16 YinstanceIdentifier
0x11	17 Yleafref
0x12	18 Ystring
0x13	19 Yunion
```

## FRR mgmtd cli

ok
```
mgmt set-config /frr-routing:routing/control-plane-protocols/control-plane-protocol[type='frr-staticd:staticd'][name='staticd'][vrf='default'] {}
```

ng
```
mgmt set-config /frr-routing:routing/control-plane-protocols/control-plane-protocol[type='staticd'][name='staticd'][vrf='default'] {}
```

full snippet
```
mgmt set-config /frr-routing:routing/control-plane-protocols/control-plane-protocol[type='frr-staticd:staticd'][name='staticd'][vrf='default'] {}
mgmt set-config /frr-routing:routing/control-plane-protocols/control-plane-protocol[type='frr-staticd:staticd'][name='staticd'][vrf='default']/frr-staticd:staticd/route-list[prefix='1.1.1.1/32'][afi-safi='frr-routing:ipv4-unicast'] {}
mgmt set-config /frr-routing:routing/control-plane-protocols/control-plane-protocol[type='frr-staticd:staticd'][name='staticd'][vrf='default']/frr-staticd:staticd/route-list[prefix='1.1.1.1/32'][afi-safi='frr-routing:ipv4-unicast']/path-list[table-id='0'][distance='1'] {}
mgmt set-config /frr-routing:routing/control-plane-protocols/control-plane-protocol[type='frr-staticd:staticd'][name='staticd'][vrf='default']/frr-staticd:staticd/route-list[prefix='1.1.1.1/32'][afi-safi='frr-routing:ipv4-unicast']/path-list[table-id='0'][distance='1']/frr-nexthops/nexthop[nh-type='blackhole'][vrf='default'][gateway=''][interface='(null)'] {}
```

## Developer Makefile

```
cat <<EOF > local.mk
YANG1 := ./pkg/vtyang/testdata/yang/basic
YANG2 := ./pkg/vtyang/testdata/yang/frr_mgmtd_minimal
YANG3 := ./pkg/vtyang/testdata/yang/choice_case
YANG := \$(YANG1)
r: vtyang-build
	sudo ./bin/vtyang agent \\
		--run /var/run/vtyang \\
		--yang \$(YANG) \\
		#END
rr: vtyang-build
	sudo ./bin/vtyang agent \\
		--run /var/run/vtyang \\
		--yang \$(YANG) \\
		-c "configure" \\
		-c "set values u08 10"
		#END
run-mgmt: vtyang-build
	sudo ./bin/vtyang agent \\
		--run /var/run/vtyang \\
		--yang \$(YANG2) \\
		--mgmtd-sock /var/run/frr/mgmtd_fe.sock \\
		#END
EOF
```

## gnmic

```
gnmic -a sandbox-iosxr-1.cisco.com:57777 \
-u admin -p C1sco12345 --insecure -e json_ietf \
--file yang/vendor/cisco/xr/732/openconfig-interfaces.yang \
--dir yang/standard/ietf \
--exclude ietf-interfaces \
get --path '/interfaces'

gnmic -a sandbox-iosxr-1.cisco.com:57777 \
-u admin -p C1sco12345 --insecure -e json_ietf \
get --path '/interfaces'

gnmic -a sandbox-iosxr-1.cisco.com:57777 \
-u admin -p C1sco12345 --insecure -e json_ietf \
get --path '/openconfig-interfaces:interfaces/interface[name="GigabitEthernet0/0/0/2"]'

gnmic -a sandbox-iosxr-1.cisco.com:57777 \
-u admin -p C1sco12345 --insecure -e json_ietf \
get --path '/Cisco-IOS-XR-um-interface-cfg:interfaces/interface[interface-name="GigabitEthernet0/0/0/2"]'

gnmic -a sandbox-iosxr-1.cisco.com:57777 \
-u admin -p C1sco12345 --insecure -e json_ietf \
get --path '/Cisco-IOS-XR-ipv4-arp-oper:arp/nodes/node[node-name="0/RP0/CPU0"]'


// clear arp-cache mgmtEth 0/RP0/CPU0/0 location all
// (rpc clear-arp-cache-interface-location)
grpcurl -plaintext \
-H "username: admin" -H "password: C1sco12345" \
-d '{"ReqId":1,"yangpathjson":"{\"Cisco-IOS-XR-ipv4-arp-act:clear-arp-cache-interface-location\":{\"node-location\":\"0/RP0/CPU0\"}}"}' \
sandbox-iosxr-1.cisco.com:57777 IOSXRExtensibleManagabilityService.gRPCExec.ActionJSON

gnmic -a sandbox-iosxr-1.cisco.com:57777 \
-u admin -p C1sco12345 --insecure -e json_ietf \
get --path '/Cisco-IOS-XR-ipv4-arp-oper:arp/nodes/node[node-name="0/RP0/CPU0"]/adjacency-history-interface-names/adjacency-history-interface-name[interface-name="MgmtEth0/RP0/CPU0/0"]/arp-entry'

grpcurl -plaintext \
-proto github.com/openconfig/gnmi/proto/gnmi/gnmi.proto \
-proto github.com/openconfig/gnmi/proto/gnmi_ext/gnmi_ext.proto \
-H "username: admin" \
-H "password: C1sco12345" \
sandbox-iosxr-1.cisco.com:57777 \
gnmi.gNMI.Capabilities

grpcurl -plaintext \
-H "username: admin" -H "password: C1sco12345" \
sandbox-iosxr-1.cisco.com:57777 list
```

## Cisco IOS-XR gRPC automation

show operational state
```
// show arp
gnmic -a sandbox-iosxr-1.cisco.com:57777 \
-u admin -p C1sco12345 --insecure -e json_ietf \
get --path '/Cisco-IOS-XR-ipv4-arp-oper:arp/nodes/node[node-name="0/RP0/CPU0"]/entries'
```

execute rpc
```
// clear arp-cache mgmtEth 0/RP0/CPU0/0 location all
grpcurl -plaintext \
-H "username: admin" -H "password: C1sco12345" \
-d '{"ReqId":1,"yangpathjson":"{\"Cisco-IOS-XR-ipv4-arp-act:clear-arp-cache-interface-location\":{\"node-location\":\"0/RP0/CPU0\"}}"}' \
sandbox-iosxr-1.cisco.com:57777 \
IOSXRExtensibleManagabilityService.gRPCExec.ActionJSON
```

get configuration
```
// show running-config arp 10.10.20.111 5254.0000.0001 ARPA
gnmic -a sandbox-iosxr-1.cisco.com:57777 \
-u admin -p C1sco12345 --insecure -e json_ietf \
get --path '/Cisco-IOS-XR-ipv4-arp-cfg:arpgmp'

gnmic -a sandbox-iosxr-1.cisco.com:57777 \
-u admin -p C1sco12345 --insecure -e json_ietf \
get --delete '/Cisco-IOS-XR-ipv4-arp-cfg:arpgmp/vrf[vrf-name="default"]/entries/entry[address="10.10.20.111"]'
```

delete configuration (gNMI)
```
gnmic -a sandbox-iosxr-1.cisco.com:57777 \
-u admin -p C1sco12345 --insecure -e json_ietf \
set --delete '/Cisco-IOS-XR-ipv4-arp-cfg:arpgmp/vrf[vrf-name="default"]/entries/entry[address="10.10.20.111"]'
```

set configuration (gNMI)
```
// configure
// arp 10.10.20.111 5254.0000.0001 ARPA interface MgmtEth 0/RP0/CPU0/0
gnmic -a sandbox-iosxr-1.cisco.com:57777 \
-u admin -p C1sco12345 --insecure -e json_ietf set \
--update-path '/Cisco-IOS-XR-ipv4-arp-cfg:arpgmp/vrf[vrf-name="default"]/entries/entry[address="10.10.20.111"]/encapsulation' \
--update-value arpa \
--update-path '/Cisco-IOS-XR-ipv4-arp-cfg:arpgmp/vrf[vrf-name="default"]/entries/entry[address="10.10.20.111"]/entry-type' \
--update-value static \
--update-path '/Cisco-IOS-XR-ipv4-arp-cfg:arpgmp/vrf[vrf-name="default"]/entries/entry[address="10.10.20.111"]/interface' \
--update-value "MgmtEth0/RP0/CPU0/0" \
--update-path '/Cisco-IOS-XR-ipv4-arp-cfg:arpgmp/vrf[vrf-name="default"]/entries/entry[address="10.10.20.111"]/mac-address' \
--update-value "52:54:00:00:00:01"
```

set configuration (cisco gRPC)
```
grpcurl -plaintext \
-H "username: admin" -H "password: C1sco12345" \
-d '{"ReqId":1,"yangjson":"{\"Cisco-IOS-XR-ipv4-arp-cfg:arpgmp\":{\"vrf\":[{\"vrf-name\":\"default\",\"entries\":{\"entry\":[{\"address\":\"10.10.20.111\",\"encapsulation\":\"arpa\",\"entry-type\":\"static\",\"interface\":\"MgmtEth0/RP0/CPU0/0\",\"mac-address\":\"52:54:00:00:00:01\"}]}}]}}"}' \
sandbox-iosxr-1.cisco.com:57777 IOSXRExtensibleManagabilityService.gRPCConfigOper.MergeConfig

grpcurl -plaintext \
-H "username: admin" -H "password: C1sco12345" \
-d '{"ReqId":1,"yangjson":"{\"Cisco-IOS-XR-ipv4-arp-cfg:arpgmp\":{\"vrf\":[{\"vrf-name\":\"default\",\"entries\":{\"entry\":[{\"address\":\"10.10.20.111\",\"encapsulation\":\"arpa\",\"entry-type\":\"static\",\"interface\":\"MgmtEth0/RP0/CPU0/0\",\"mac-address\":\"52:54:00:00:00:01\"}]}}]}}"}' \
sandbox-iosxr-1.cisco.com:57777 IOSXRExtensibleManagabilityService.gRPCConfigOper.ReplaceConfig
```

## Presentation Materials

- todo
