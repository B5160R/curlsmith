.PHONY: build test cover lint clean

build:
	go build ./...

test:
	go test ./... -count=1

cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out
	@echo ""
	@echo "To view in browser: go tool cover -html=coverage.out"

lint:
	go vet ./...

clean:
	rm -f coverage.out
