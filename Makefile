test:
	go test ./...
godoc:
	godoc -http=:6060
generate:
	go generate ./...
include ./cmd/*/sub.mk
