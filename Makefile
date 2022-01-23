build:
	go build -o vtyang main.go 
run: build
	./vtyang agent
r: build
	./vtyang -p ./yang
kill:
	killall vtyang
