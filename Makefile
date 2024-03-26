BIN_NAME=bivrost
TEST_BIN_NAME=testmodule

WINDOWS=$(BIN_NAME)_win_amd64.exe
LINUX=$(BIN_NAME)_linux_amd64.out

TEST_WINDOWS=$(TEST_BIN_NAME)_win_amd64.exe
TEST_LINUX=$(TEST_BIN_NAME)_linux_amd64.out

VERSION=$(shell git describe --tags --always --long)

.PHONY: all test clean

$(LINUX): cmd/bivrost/main.go
	CGO_ENABLED=1 go build -v -o $(LINUX) -tags linux -ldflags="-s -w -X main.version=$(VERSION)" ./cmd/bivrost/main.go

$(TEST_LINUX): cmd/testmodule/main.go
	CGO_ENABLED=1 go build -v -o $(TEST_LINUX) -tags linux -ldflags="-s -w -X main.version=$(VERSION)" ./cmd/testmodule/main.go

$(WINDOWS): cmd/bivrost/main.go
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -v -o $(WINDOWS) -ldflags="-s -w -X main.version=$(VERSION)" ./cmd/bivrost/main.go

$(TEST_WINDOWS): cmd/testmodule/main.go
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -v -o $(TEST_WINDOWS) -ldflags="-s -w -X main.version=$(VERSION)" ./cmd/testmodule/main.go

# Build targets
windows: $(WINDOWS)
linux: $(LINUX)
prototype: $(TEST_LINUX) $(LINUX)

test: go test ./...

build: windows linux
	@echo $(VERSION)
	@echo "Build complete"

run: $(LINUX) && ./$(LINUX)

run-prototype: # Run prototype
	$(TEST_LINUX)
	./$(TEST_LINUX)

clean:	## Remove build files
	go clean
	rm $(WINDOWS) $(LINUX) $(TEST_WINDOWS) $(TEST_LINUX)

help: ## Display available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
