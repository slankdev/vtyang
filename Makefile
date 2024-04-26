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
		-c "xpath-show lib interface dum0 description dum0-comment" \
		#END
