test:
	go test ./... -count=1
godoc:
	godoc -http=:6060
generate:
	go generate ./...
include ./cmd/*/sub.mk

YANG := ./pkg/vtyang/testdata/yang/leaf_types
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
		-c "set values u8 10"
		#END
