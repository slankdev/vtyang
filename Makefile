test:
	go test ./... -count=1 -v
godoc:
	godoc -http=:6060
generate:
	go generate ./...
include ./cmd/*/sub.mk
