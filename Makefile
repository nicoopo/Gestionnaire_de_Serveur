APP_NAME := server-manager
BUILD_DIR := ./bin
MAIN_PATH := ./cmd/server
DOCKER_COMPOSE := docker compose

.PHONY: all build run test clean fmt vet lint tidy docker-build docker-up docker-down docker-logs docker-restart help

all: build

## build: compile le binaire en local
build:
	@echo "Building $(APP_NAME)..."
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)

## run: build puis lance le binaire
run: build
	@echo "Running $(APP_NAME)..."
	$(BUILD_DIR)/$(APP_NAME)

## test: lance les tests avec la couverture
test:
	go test -v -race -cover ./...

## fmt: formate le code
fmt:
	go fmt ./...

## vet: analyse statique du code
vet:
	go vet ./...

## lint: fmt + vet regroupés
lint: fmt vet

## tidy: nettoie go.mod / go.sum
tidy:
	go mod tidy

## clean: supprime les binaires générés
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)

## docker-build: build l'image via docker compose
docker-build:
	$(DOCKER_COMPOSE) build

## docker-up: démarre les conteneurs en arrière-plan
docker-up:
	$(DOCKER_COMPOSE) up -d

## docker-down: arrête et supprime les conteneurs
docker-down:
	$(DOCKER_COMPOSE) down

## docker-restart: down puis up
docker-restart: docker-down docker-up

## docker-logs: suit les logs du service
docker-logs:
	$(DOCKER_COMPOSE) logs -f

## help: affiche cette aide
help:
	@echo "Cibles disponibles :"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'