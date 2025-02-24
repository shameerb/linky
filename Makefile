.PHONY: build install run clean dev prod docker-build docker-run

# Default markdown directory if not set
MARKDOWN_DIR ?= $(PWD)/markdown
BINARY_NAME=mde
GO_PACKAGE=./backend

# Frontend commands
frontend-install:
	cd frontend && npm install

frontend-build: frontend-install
	cd frontend && npm run build

frontend-dev:
	cd frontend && npm run dev

# Backend commands
backend-build: frontend-build
	# cd backend && mkdir -p static && cp -r ../frontend/dist/* static/
	cd backend && GO_ENV=production go build -o bin/$(BINARY_NAME) main.go

backend-dev:
	cd backend && MARKDOWN_DIR=$(MARKDOWN_DIR) go run main.go

backend-prod:
	cd backend && GO_ENV=production MARKDOWN_DIR=$(MARKDOWN_DIR) go run main.go

# Development mode
dev:
	make frontend-install
	@echo "Starting development servers..."
	@echo "Frontend will be available at http://localhost:3000"
	@echo "Backend API will be available at http://localhost:8080"
	@(trap 'kill 0' SIGINT; make backend-run-dev & make frontend-dev)

# Production mode
prod: backend-build
	@echo "Starting production server at http://localhost:8080"
	cd backend && GO_ENV=production MARKDOWN_DIR=$(MARKDOWN_DIR) ./bin/$(BINARY_NAME)

# Install binary using Go's standard approach
install: frontend-build
	# cd backend && mkdir -p static && cp -r ../frontend/dist/* static/
	cd backend && GO_ENV=production go install
	@echo "Installed $(BINARY_NAME) using go install"
	@echo "Make sure your Go bin directory is in your PATH"

clean:
	rm -rf backend/bin
	rm -rf backend/dist
	rm -rf frontend/node_modules
	rm -rf frontend/dist
	# rm -rf backend/static

# Docker commands
docker-build:
	docker build -t linky:latest .

docker-run:
	mkdir -p $(MARKDOWN_DIR)
	docker run -p 8080:8080 -v $(MARKDOWN_DIR):/data/markdown linky:latest
