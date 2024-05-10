test:
	go test ./... -count=1 -v
test-summary:
	go test ./... -count=1 -v -json \
		| jq '. | select(.Action == "fail") | select(.Test != null) | .Test' -r
godoc:
	godoc -http=:6060
generate:
	go generate ./...
include ./cmd/*/sub.mk
-include local.mk
