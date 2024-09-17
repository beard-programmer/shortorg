# Name of the binary that will be built
BINARY_NAME=shortorg

# Main file for the Go application
MAIN=main.go

# Go flags
GOFLAGS=

# Default target, compiles the code
.PHONY: all
all: build

# Build the binary
.PHONY: build
build:
	@echo "Building the application..."
	GOFLAGS=$(GOFLAGS) go build -o $(BINARY_NAME) $(MAIN)

# Run the application
.PHONY: run
run:
	@echo "Running the application..."
	GOFLAGS=$(GOFLAGS) go run $(MAIN)

# Clean up the build artifacts
.PHONY: clean
clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)

# Format the code with gofmt
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	gofmt -w .

# Tidy the Go modules
.PHONY: tidy
tidy:
	@echo "Tidying Go modules..."
	go mod tidy

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test ./...

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod download

# Lint the code
.PHONY: lint
lint:
	@echo "Linting code..."
	golangci-lint run
