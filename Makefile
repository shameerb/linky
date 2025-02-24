.PHONY: build run clean dev prod docker-build docker-run

# Default markdown directory if not set
MARKDOWN_DIR ?= $(PWD)/markdown

# Frontend commands
frontend-install:
	cd frontend && npm install

frontend-build: frontend-install
	cd frontend && npm run build

frontend-dev:
	cd frontend && npm run dev

# Backend commands
backend-build:
	cd backend && go build -o bin/server main.go

backend-run-dev:
	cd backend && MARKDOWN_DIR=$(MARKDOWN_DIR) go run main.go

backend-run-prod:
	cd backend && GO_ENV=production MARKDOWN_DIR=$(MARKDOWN_DIR) go run main.go

# Development mode
dev:
	make frontend-install
	@echo "Starting development servers..."
	@echo "Frontend will be available at http://localhost:3000"
	@echo "Backend API will be available at http://localhost:8080"
	@(trap 'kill 0' SIGINT; make backend-run-dev & make frontend-dev)

# Production mode
prod: frontend-build backend-build
	@echo "Starting production server at http://localhost:8080"
	cd backend && GO_ENV=production MARKDOWN_DIR=$(MARKDOWN_DIR) ./bin/server

clean:
	rm -rf backend/bin
	rm -rf backend/dist
	rm -rf frontend/node_modules
	rm -rf frontend/dist

# Docker commands
docker-build:
	docker build -t linky:latest .

docker-run:
	mkdir -p $(MARKDOWN_DIR)
	docker run -p 8080:8080 -v $(MARKDOWN_DIR):/data/markdown linky:latest
