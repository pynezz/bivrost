BIN_NAME=bivrost
TEST_BIN_NAME=testmodule

WINDOWS=$(BIN_NAME)_win_amd64.exe
LINUX=$(BIN_NAME)_linux_amd64.out

TEST_WINDOWS=$(TEST_BIN_NAME)_win_amd64.exe
TEST_LINUX=$(TEST_BIN_NAME)_linux_amd64.out

VERSION=$(shell git describe --tags --always --long)

.PHONY: all test clean

# Build targets
windows: $(WINDOWS)
linux: $(LINUX)



$(LINUX):
	CGO_ENABLED=1 go build -v -o $(LINUX) -tags linux -ldflags="-s -w -X main.version=$(VERSION)" ./cmd/bivrost/main.go
$(TEST_LINUX):
	CGO_ENABLED=1 go build -v -o $(TEST_LINUX) -tags linux -ldflags="-s -w -X main.version=$(VERSION)" ./cmd/testmodule/main.go

$(WINDOWS): GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -v -o $(WINDOWS) -ldflags="-s -w -X main.version=$(VERSION)" ./cmd/bivrost/main.go
$(TEST_WINDOWS): GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -v -o $(TEST_WINDOWS) -ldflags="-s -w -X main.version=$(VERSION)" ./cmd/testmodule/main.go

test:
	go test ./...

prototype: $(TEST_LINUX) $(LINUX)


build: windows linux
	@echo $(VERSION)
	@echo "Build complete"

run: $(LINUX) && ./$(LINUX)

run-prototype: $(TEST_LINUX)
	./$(TEST_LINUX)

clean:
	go clean
	rm $(WINDOWS) $(LINUX) $(TEST_WINDOWS) $(TEST_LINUX)
