test:
	go test ./... -count=1
godoc:
	godoc -http=:6060
generate:
	go generate ./...
include ./cmd/*/sub.mk

r: vtyang-build
	sudo ./bin/vtyang agent \
		--run-path /var/run/vtyang \
		--yang ./pkg/vtyang/testdata/yang/frr_mgmtd_minimal2 \
		#END
rr: vtyang-build
	sudo ./bin/vtyang agent \
		--run-path /var/run/vtyang \
		--yang ./pkg/vtyang/testdata/yang/frr_mgmtd_minimal2 \
		-c "configure" \
		-c "set lib prefix-list ipv4 hoge entry 10 action permit" \
		#-c "show-xpath lib prefix-list ipv4 hoge entry 10 action permit" \
		#END
