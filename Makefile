# Variables
BINARY_NAME := kubelitedb
BUILD_DIR := build
KUBECONFIG := "${HOME}/.kube/config"

# Targets
.PHONY: all build deploy clean

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) .

clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)

# Testing
kind-up:
	@echo "Creating a kind cluster..."
	@kind create cluster --name kubelitedb || true

kind-down:
	@echo "Deleting the kind cluster..."
	@kind delete cluster --name kubelitedb || true
	
deploy:
	@echo "Deploying $(BINARY_NAME) to Kubernetes..."
	@kubectl apply -f crds/

up: kind-up deploy build
	@echo "Starting the controller..."
	@$(BUILD_DIR)/$(BINARY_NAME) -kubeconfig=$(KUBECONFIG)

debug: kind-up deploy
	@echo "Starting the controller in debug mode..."
	@dlv debug --headless --listen=:2345 --api-version=2 --accept-multiclient . -- -kubeconfig=$(KUBECONFIG)

install:
	sudo cp ./usr/local/bin/kubectl-sqlite /usr/local/bin/kubectl-sqlite