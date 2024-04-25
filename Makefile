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
	sudo rm -f /tmp/vtyang.log
	sudo ./bin/vtyang agent --run-path /var/run/vtyang -y ./yang.frr2
r:
	GOOS=linux CGO_ENABLED=0 go build -o bin/linux-agent cmd/linux-agent/main.go
	scp ./bin/linux-agent dev:/tmp
	ssh -t dev sudo /tmp/linux-agent
reset-netns:
	ssh -t dev sudo ip netns del ns0 || true
	ssh -t dev sudo ip netns add ns0
	ssh -t dev sudo ip netns exec ns0 ip link add dum0 type dummy
	ssh -t dev sudo ip netns exec ns0 ip link add dum1 type dummy
	ssh -t dev sudo ip netns exec ns0 ip link add dum2 type dummy
frr-run: frr-agent-build
	ssh -t dev sudo /home/ubuntu/git/vtyang/bin/frr-agent

t:
	sudo rm -f /tmp/vtyang.log
	go test ./pkg/vtyang/ -count=1 -run TestYangCompletion1
tt: vtyang-build
	./bin/vtyang agent -y pkg/vtyang/testdata/same_container_name_in_different_modules/
