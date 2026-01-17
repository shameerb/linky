.PHONY: build run dev clean frontend-dev backend-dev install

# Build settings
BINARY_NAME=linky
WEB_DIR=web
STATIC_DIR=internal/static

# Installation settings
GOBIN ?= $(shell go env GOPATH)/bin

# Development settings
DEV_PORT=8080
MARKDOWN_DIR=./Links

build: build-frontend build-backend

build-frontend:
	@echo "Cleaning old frontend builds..."
	rm -rf internal/static/dist/*
	rm -rf $(WEB_DIR)/dist/*
	@echo "Creating dist directory..."
	mkdir -p internal/static/dist
	@echo "Building frontend..."
	cd $(WEB_DIR) && npm install && npm run build
	@echo "Frontend built successfully to internal/static/dist"

build-backend:
	@echo "Building backend..."
	mkdir -p bin
	go build -o bin/$(BINARY_NAME) cmd/server/main.go

# Install the binary to Go bin directory
install: build
	@echo "Installing $(BINARY_NAME) to $(GOBIN)..."
	cp bin/$(BINARY_NAME) $(GOBIN)/
	@echo "Installation complete. Make sure $(GOBIN) is in your PATH"
	@echo "You can now run '$(BINARY_NAME)' from anywhere"

run: build
	@echo "Running server..."
	MARKDOWN_DIR=$(MARKDOWN_DIR) ./bin/$(BINARY_NAME)

# Development commands
frontend-dev:
	@echo "Starting frontend development server..."
	cd $(WEB_DIR) && npm run dev

backend-dev:
	@echo "Starting backend development server..."
	MARKDOWN_DIR=$(MARKDOWN_DIR) go run cmd/server/main.go

# Run both frontend and backend in development mode
dev:
	@echo "Starting development servers..."
	@echo "Frontend will be available at http://localhost:8080"
	@echo "Backend API will be available at http://localhost:8080"
	@(trap 'kill 0' SIGINT; make backend-dev & make frontend-dev)

clean:
	@echo "Cleaning up..."
	rm -rf bin
	rm -rf internal/static/dist
	rm -rf $(WEB_DIR)/dist
	rm -rf $(WEB_DIR)/node_modules
