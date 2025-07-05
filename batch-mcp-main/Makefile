.PHONY: build dist test clean help


# Go project variables
BINARY_NAME=batch-mcp
BUILD_DIR=build
SRC_DIR=./cmd/batch-mcp
RELEASE_SCRIPT=release.sh

# Build the binary


# Build the binary for the current platform
build: | $(BUILD_DIR)
	go mod tidy
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(SRC_DIR)


# Build the binary for linux amd64
build-linux-amd64: | $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(SRC_DIR)

dist: build
	./$(RELEASE_SCRIPT)

# Run tests
test:
	go test ./... -v

# Clean up build artifacts
clean:
	go clean
	rm -rf $(BUILD_DIR)
	rm -rf ./dist/
# Create build directory if it doesn't exist
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

# Display help
help:
	@echo "Makefile for Go project"
	@echo "Usage:"
	@echo "  make build   - Build the binary"
	@echo "  make dist    - Build distribution binaries for multiple platforms"
	@echo "  make test    - Run tests"
	@echo "  make clean   - Clean up build artifacts"
	@echo "  make help    - Display this help message"
