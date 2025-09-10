.PHONY: help build run test clean docker-build docker-run migrate-up migrate-down docker-restart docker-stop start-all stop-all start-ollama stop-ollama build-images build-images-app build-images-docreader build-images-frontend clean-images

# Show help
help:
	@echo "WeKnowRust Makefile Help"
	@echo ""
	@echo "Basic Commands:"
	@echo "  build             Build the application"
	@echo "  run               Run the application"
	@echo "  test              Run tests"
	@echo "  clean             Clean build artifacts"
	@echo ""
	@echo "Docker Commands:"
	@echo "  docker-build      Build Docker image"
	@echo "  docker-run        Run Docker containers"
	@echo "  docker-stop       Stop Docker containers"
	@echo "  docker-restart    Restart Docker containers"
	@echo ""
	@echo "Service Management:"
	@echo "  start-all         Start all services"
	@echo "  stop-all          Stop all services"
	@echo "  start-ollama      Start only the Ollama service"
	@echo ""
	@echo "Image Build:"
	@echo "  build-images      Build all images from source"
	@echo "  build-images-app  Build the app image from source"
	@echo "  build-images-docreader Build the doc reader image from source"
	@echo "  build-images-frontend  Build the frontend image from source"
	@echo "  clean-images      Clean local images"
	@echo ""
	@echo "Database:"
	@echo "  migrate-up        Apply database migrations"
	@echo "  migrate-down      Roll back database migrations"
	@echo ""
	@echo "Dev Tools:"
	@echo "  fmt               Format code"
	@echo "  lint              Lint code"
	@echo "  deps              Install dependencies"
	@echo "  docs              Generate API documentation"

# Go related variables
BINARY_NAME=WeKnowRust
MAIN_PATH=./cmd/server

# Docker related variables
DOCKER_IMAGE=WeKnowRust
DOCKER_TAG=latest

# Build the application
build:
	go build -o $(BINARY_NAME) $(MAIN_PATH)

# Run the application
run: build
	./$(BINARY_NAME)

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	go clean
	rm -f $(BINARY_NAME)

# Build Docker image
docker-build:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

# Run Docker container (传统方式)
docker-run:
	docker-compose up

# 使用新脚本启动所有服务
start-all:
	./scripts/start_all.sh

# 使用新脚本仅启动Ollama服务
start-ollama:
	./scripts/start_all.sh --ollama

# 使用新脚本仅启动Docker容器
start-docker:
	./scripts/start_all.sh --docker

# 使用新脚本停止所有服务
stop-all:
	./scripts/start_all.sh --stop

# Stop Docker container (传统方式)
docker-stop:
	docker-compose down

# 从源码构建镜像相关命令
build-images:
	./scripts/build_images.sh

build-images-app:
	./scripts/build_images.sh --app

build-images-docreader:
	./scripts/build_images.sh --docreader

build-images-frontend:
	./scripts/build_images.sh --frontend

clean-images:
	./scripts/build_images.sh --clean

# Restart Docker container (stop, rebuild, start)
docker-restart:
	docker-compose stop -t 60
	docker-compose up --build

# Database migrations
migrate-up:
	./scripts/migrate.sh up

migrate-down:
	./scripts/migrate.sh down

# Generate API documentation
docs:
	swag init -g $(MAIN_PATH)/main.go -o ./docs

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Install dependencies
deps:
	go mod download

# Build for production
build-prod:
	GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o $(BINARY_NAME) $(MAIN_PATH)

clean-db:
	@echo "Cleaning database..."
	@if [ $$(docker volume ls -q -f name=weknowrust_postgres-data) ]; then \
		docker volume rm weknowrust_postgres-data; \
	fi
	@if [ $$(docker volume ls -q -f name=weknowrust_minio_data) ]; then \
		docker volume rm weknowrust_minio_data; \
	fi
	@if [ $$(docker volume ls -q -f name=weknowrust_redis_data) ]; then \
		docker volume rm weknowrust_redis_data; \
	fi


