.PHONY: test
.SILENT:

install:
	@echo "=== Installing dependencies ==="
	@dep ensure -v

mocks:
	@echo "=== Creating mocks ==="
	rm -fr mocks
	CGO_ENABLED=0 mockery -name . -inpkg

test: mocks
	@echo "=== Running Unit Tests ==="
	go test -race -coverprofile=.coverage.out  ./...
	go tool cover -func=.coverage.out