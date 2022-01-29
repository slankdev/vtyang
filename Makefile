build:
	go build -o vtyang main.go
test:
	go test ./...
run: build
	./vtyang agent --dbpath ./tmp/config.json --run-path /usr/local/var/run/vtyang
kill:
	killall vtyang
log:
	tail -F /tmp/vtyang.log
watch:
	@cat ./tmp/config.json
	@echo
	@echo
	@ls -l /usr/local/var/run/vtyang
