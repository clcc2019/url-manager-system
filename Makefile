# URL Manager System Makefile

.PHONY: help build test clean dev deploy

# Variables
BACKEND_IMAGE = url-manager/backend
FRONTEND_IMAGE = url-manager/frontend
VERSION = latest

help: ## Show this help
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development
dev: ## Start development environment
	cd deployments/docker && sudo docker compose -f docker-compose.dev.yml up -d

dev-local: ## Start local development environment (no containers for backend/frontend)
	cd deployments/docker && sudo docker compose -f docker-compose.dev.yml up -d postgres redis
	@echo "Database services started!"
	@echo "PostgreSQL: localhost:5432"
	@echo "Redis: localhost:6379"
	@echo "Run 'make dev-backend-local' and 'make dev-frontend-local' in separate terminals"

dev-backend-local: ## Start backend in local development mode
	pkill main || true
	cd backend && go run cmd/server/main.go

dev-frontend-local: ## Start frontend in local development mode
	cd frontend && npm run dev -- --host 0.0.0.0 --port 5173
	@echo "Development environment started!"
	@echo "Backend: http://localhost:8080"
	@echo "Frontend: http://localhost:5173"
	@echo "PostgreSQL: localhost:5432"
	@echo "Redis: localhost:6379"

dev-logs: ## Show development logs
	cd deployments/docker && sudo docker compose -f docker-compose.dev.yml logs -f

dev-stop: ## Stop development environment
	cd deployments/docker && sudo docker compose -f docker-compose.dev.yml down

dev-clean: ## Clean development environment and volumes
	cd deployments/docker && sudo docker compose -f docker-compose.dev.yml down -v

# Building
build: build-backend build-frontend ## Build all Docker images

build-backend: ## Build backend Docker image
	sudo docker build -f backend/Dockerfile -t $(BACKEND_IMAGE):$(VERSION) .

build-frontend: ## Build frontend Docker image
	sudo docker build -f frontend/Dockerfile -t $(FRONTEND_IMAGE):$(VERSION) .

build-local: ## Build images for local development
	./scripts/build-local.sh build

build-and-test: ## Build and test images locally
	./scripts/build-local.sh all

# Testing
test: test-backend test-frontend ## Run all tests

test-backend: ## Run backend tests
	cd backend && go test ./...

test-frontend: ## Run frontend tests
	cd frontend && npm test

# Linting
lint: lint-backend lint-frontend ## Run all linting

lint-backend: ## Run backend linting
	cd backend && golangci-lint run

lint-frontend: ## Run frontend linting
	cd frontend && npm run lint

# Database
db-migrate: ## Run database migrations
	cd backend && go run cmd/migrate/main.go

db-seed: ## Seed database with sample data
	cd backend && go run cmd/seed/main.go

# Kubernetes
k8s-deploy: ## Deploy to Kubernetes using Helm
	helm upgrade --install url-manager deployments/helm/url-manager \
		--create-namespace \
		--namespace url-manager \
		--wait

k8s-uninstall: ## Uninstall from Kubernetes
	helm uninstall url-manager --namespace url-manager

k8s-template: ## Show Kubernetes templates
	helm template url-manager deployments/helm/url-manager

# Production
prod: ## Start production environment
	cd deployments/docker && sudo docker compose up -d

prod-stop: ## Stop production environment
	cd deployments/docker && sudo docker compose down

# Cleanup
clean: ## Clean up Docker images and containers
	sudo docker system prune -f
	sudo docker volume prune -f

# Security
security-scan: ## Run security scans
	@echo "Running backend security scan..."
	cd backend && gosec ./...
	@echo "Running frontend security scan..."
	cd frontend && npm audit

# Documentation
docs: ## Generate documentation
	@echo "Generating API documentation..."
	cd backend && swag init -g cmd/server/main.go

# Monitoring
monitor: ## Start monitoring stack (Prometheus + Grafana)
	cd deployments/monitoring && sudo docker compose up -d

monitor-stop: ## Stop monitoring stack
	cd deployments/monitoring && sudo docker compose down