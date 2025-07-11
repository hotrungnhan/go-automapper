.PHONY: lint

lint:
	go vet ./...
	golangci-lint run
	
benchmark:
	go test -bench=. ./...

.PHONY: test test-all

test:
	go test ./...

coverage:
	go test -v ./... -cover