build:
	go build -o vtyang main.go
test:
	go test ./...
run: build
	./vtyang agent --dbpath ./tmp/config.json
kill:
	killall vtyang
log:
	tail -F /tmp/vtyang.log
