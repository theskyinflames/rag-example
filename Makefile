# Makefile for RAG Example Application

# Variables
IMAGE_NAME := rag-example
CONTAINER_NAME := rag-app

# Default target
.PHONY: help
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Build targets
.PHONY: build
build: ## Build the Docker image
	@echo "Building Docker image..."
	docker build -t $(IMAGE_NAME) .

.PHONY: build-no-cache
build-no-cache: ## Build the Docker image without cache
	@echo "Building Docker image without cache..."
	docker build --no-cache -t $(IMAGE_NAME) .

# Run targets
.PHONY: run
run: ## Run the container with Docker
	@echo "Running container..."
	@if [ -z "$(OPENAI_API_KEY)" ]; then \
		echo "Error: OPENAI_API_KEY environment variable is not set"; \
		echo "Please set it with: export OPENAI_API_KEY='your-api-key'"; \
		exit 1; \
	fi
	docker run --rm --name $(CONTAINER_NAME) \
		-e OPENAI_API_KEY=$(OPENAI_API_KEY) \
		$(IMAGE_NAME)

.PHONY: run-interactive
run-interactive: ## Run the container interactively
	@echo "Running container interactively..."
	@if [ -z "$(OPENAI_API_KEY)" ]; then \
		echo "Error: OPENAI_API_KEY environment variable is not set"; \
		echo "Please set it with: export OPENAI_API_KEY='your-api-key'"; \
		exit 1; \
	fi
	docker run --rm -it --name $(CONTAINER_NAME) \
		-e OPENAI_API_KEY=$(OPENAI_API_KEY) \
		$(IMAGE_NAME)

.PHONY: run-detached
run-detached: ## Run the container in detached mode
	@echo "Running container in detached mode..."
	@if [ -z "$(OPENAI_API_KEY)" ]; then \
		echo "Error: OPENAI_API_KEY environment variable is not set"; \
		echo "Please set it with: export OPENAI_API_KEY='your-api-key'"; \
		exit 1; \
	fi
	docker run -d --name $(CONTAINER_NAME) \
		-e OPENAI_API_KEY=$(OPENAI_API_KEY) \
		$(IMAGE_NAME)

# Development targets
.PHONY: dev
dev: ## Run the application locally (requires Go)
	@echo "Running application locally..."
	@if [ -z "$(OPENAI_API_KEY)" ]; then \
		echo "Error: OPENAI_API_KEY environment variable is not set"; \
		echo "Please set it with: export OPENAI_API_KEY='your-api-key'"; \
		exit 1; \
	fi
	go run main.go

.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	go test ./...

.PHONY: mod-tidy
mod-tidy: ## Tidy Go modules
	@echo "Tidying Go modules..."
	go mod tidy

# Cleanup targets
.PHONY: clean
clean: ## Remove the Docker image
	@echo "Removing Docker image..."
	docker rmi $(IMAGE_NAME) 2>/dev/null || true

.PHONY: clean-all
clean-all: ## Remove containers, images, and volumes
	@echo "Cleaning up all Docker resources..."
	docker rmi $(IMAGE_NAME) 2>/dev/null || true
	docker system prune -f

.PHONY: stop
stop: ## Stop running containers
	@echo "Stopping containers..."
	docker stop $(CONTAINER_NAME) 2>/dev/null || true

# Utility targets
.PHONY: logs
logs: ## Show logs from the running container
	docker logs -f $(CONTAINER_NAME)

.PHONY: shell
shell: ## Get a shell in the running container
	docker exec -it $(CONTAINER_NAME) /bin/sh

.PHONY: inspect
inspect: ## Inspect the Docker image
	docker inspect $(IMAGE_NAME)

.PHONY: size
size: ## Show the size of the Docker image
	docker images $(IMAGE_NAME) --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}"

# Quick start targets
.PHONY: quick-start
quick-start: build run ## Build and run the application quickly

# Check environment
.PHONY: check-env
check-env: ## Check if required environment variables are set
	@echo "Checking environment variables..."
	@if [ -z "$(OPENAI_API_KEY)" ]; then \
		echo "❌ OPENAI_API_KEY is not set"; \
		echo "   Please set it with: export OPENAI_API_KEY='your-api-key'"; \
	else \
		echo "✅ OPENAI_API_KEY is set"; \
	fi
	@echo "✅ Environment check complete"
