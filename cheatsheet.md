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

grpcurl -plaintext \
-proto github.com/openconfig/gnmi/proto/gnmi/gnmi.proto \
-proto github.com/openconfig/gnmi/proto/gnmi_ext/gnmi_ext.proto \
-H "username: admin" \
-H "password: C1sco12345" \
sandbox-iosxr-1.cisco.com:57777 \
gnmi.gNMI.Capabilities
```

## Presentation Materials

- todo
