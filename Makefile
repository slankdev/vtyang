test:
	go test ./... -count=1
godoc:
	godoc -http=:6060
generate:
	go generate ./...
include ./cmd/*/sub.mk

YANG1 := ./pkg/vtyang/testdata/yang/leaf_types
YANG2 := ./pkg/vtyang/testdata/yang/frr_mgmtd_minimal
YANG := $(YANG1)
r: vtyang-build
	sudo ./bin/vtyang agent \
		--run-path /var/run/vtyang \
		--yang $(YANG) \
		#END
rr: vtyang-build
	sudo ./bin/vtyang agent \
		--run-path /var/run/vtyang \
		--yang $(YANG) \
		-c "configure" \
		-c "set values u08 10"
		#END
run-mgmt: vtyang-build
	sudo ./bin/vtyang agent \
		--run-path /var/run/vtyang \
		--yang $(YANG2) \
		#END
