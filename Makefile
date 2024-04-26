test:
	go test ./... -count=1
godoc:
	godoc -http=:6060
generate:
	go generate ./...
include ./cmd/*/sub.mk
rrr: vtyang-build
	./bin/vtyang agent --run-path /usr/local/var/run/vtyang -y ./yang
rr: vtyang-build
	sudo ./bin/vtyang agent \
		--run-path /var/run/vtyang \
		--yang ./pkg/vtyang/testdata/yang/frr_mgmtd_minimal2
reset-netns:
	ssh -t dev sudo ip netns del ns0 || true
	ssh -t dev sudo ip netns add ns0
	ssh -t dev sudo ip netns exec ns0 ip link add dum0 type dummy
	ssh -t dev sudo ip netns exec ns0 ip link add dum1 type dummy
	ssh -t dev sudo ip netns exec ns0 ip link add dum2 type dummy
frr-run: frr-agent-build
	ssh -t dev sudo /home/ubuntu/git/vtyang/bin/frr-agent
