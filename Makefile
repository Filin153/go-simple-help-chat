.PHONY: fix fmt vet test check
test:
	go test ./... -race -cover

fmt:
	go fmt ./...

vet:
	go vet ./...

check: fmt vet test

fix:
	go fix ./...