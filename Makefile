build:
	go build -o vtyang main.go 
run: build
	./vtyang agent
r: build
	./vtyang -p ./yang
kill:
	killall vtyang
log:
	tail -F /tmp/vtyang.log
