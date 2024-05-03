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

## Presentation Materials

- todo
