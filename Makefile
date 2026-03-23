.PHONY: run server ui agents stop docker-build deploy setup-local teardown-local port-forward

include ./.env
export

PID_DIR := /tmp/divinity
LOG_FILE := /tmp/divinity.log
AGENT_DIR := $(PID_DIR)/agents

SERVER_DIR := ./server
CLIENT_DIR := ./client
WEBSITE_DIR := ./website
IMAGE_NAME := divinity-server
KIND_CLUSTER := divinity

TEST_DIR := ./tests

IMAGE_VERSION := $(shell head -c 8 /dev/urandom | od -An -tx1 | tr -d ' \n')

# Run the server, client, and website landing page
run: stop
	@mkdir -p $(PID_DIR)
	@echo "Building Go backend ..."
	@cd $(SERVER_DIR) && go build -o $(PID_DIR)/server-bin .
	@echo "Starting Go backend on :8080 ..."
	@rm -f $(LOG_FILE)
	@echo "Output is also being written to $(LOG_FILE)"
	@$(PID_DIR)/server-bin 2>&1 | tee -a $(LOG_FILE) & echo $$! > $(PID_DIR)/server.pid
	@sleep 1
	@echo "Starting client dev server ..."
	@cd $(CLIENT_DIR) && npx vite --port 3000 2>&1 | tee -a $(LOG_FILE) & echo $$! > $(PID_DIR)/ui.pid
	@echo "Starting website landing page server ..."
	@cd $(WEBSITE_DIR) && node server.js 2>&1 | tee -a $(LOG_FILE) & echo $$! > $(PID_DIR)/landing.pid
	@echo ""
	@echo "  Server  → http://localhost:8080"
	@echo "  Client → http://localhost:3000"
	@echo "  Website  → http://localhost:3001"
	@echo ""
	@echo "Spawning NPC agents in background (cascaded) ..."
	@bash $(TEST_DIR)/spawn-agents.sh 2>&1 | tee -a $(LOG_FILE) & echo $$! > $(PID_DIR)/spawn.pid
	@echo "Press Ctrl+C or run 'make stop' to shut down."
	@wait

# Run only the Go backend
server:
	cd $(SERVER_DIR) && go run .

# Run the client and website landing page
ui:
	@cd $(CLIENT_DIR) && npx vite --port 3000 &
	@cd $(WEBSITE_DIR) && node server.js &
	@echo ""
	@echo "  Client  → http://localhost:3000"
	@echo "  Website → http://localhost:3001"
	@echo ""
	@wait

# Spawn NPC agents (standalone — server must already be running)
agents:
	bash $(TEST_DIR)/spawn-agents.sh

# Kill processes using stored PIDs, then free the ports
stop:
	@for f in $(AGENT_DIR)/*.pid; do \
		if [ -f "$$f" ]; then \
			pid=$$(cat "$$f"); \
			kill $$pid 2>/dev/null || true; \
			rm -f "$$f"; \
		fi; \
	done
	@rm -rf $(AGENT_DIR)
	@for f in $(PID_DIR)/*.pid; do \
		if [ -f "$$f" ]; then \
			pid=$$(cat "$$f"); \
			kill $$pid 2>/dev/null || true; \
			rm -f "$$f"; \
		fi; \
	done
	@-fuser -k 8080/tcp 2>/dev/null || true
	@-fuser -k 3000/tcp 2>/dev/null || true
	@-fuser -k 3001/tcp 2>/dev/null || true
	@rm -f $(PID_DIR)/server-bin $(PID_DIR)/npc-bin

# Build the Docker image for the server service
docker-build:
	docker build -t $(IMAGE_NAME):$(IMAGE_VERSION) $(SERVER_DIR)

# Deploy to Kubernetes (builds image first)
deploy: docker-build setup-local
	kubectl apply -f k8s/namespace.yaml
	kubectl apply -f k8s/mongodb.yaml
	kubectl rollout status statefulset/mongodb -n divinity --timeout=120s
	@kubectl delete secret divinity-secrets -n divinity 2>/dev/null || true
	kubectl create secret generic divinity-secrets --from-literal=OPENROUTER_KEY=$(OPENROUTER_KEY) -n divinity
	kubectl apply -f k8s/server.yaml
	kubectl set image deployment/divinity-server divinity-server=$(IMAGE_NAME):$(IMAGE_VERSION) -n divinity
	kubectl rollout status deployment/divinity-server -n divinity --timeout=120s

# Create a local Kind cluster for dev and load the server image
setup-local:
	@if kind get clusters 2>/dev/null | grep -q '^$(KIND_CLUSTER)$$'; then \
		echo "Kind cluster '$(KIND_CLUSTER)' already exists, skipping creation."; \
	else \
		echo "Creating Kind cluster '$(KIND_CLUSTER)' ..."; \
		kind create cluster --name $(KIND_CLUSTER) --wait 60s; \
	fi
	@echo "Loading image $(IMAGE_NAME):$(IMAGE_VERSION) into cluster ..."
	kind load docker-image $(IMAGE_NAME):$(IMAGE_VERSION) --name $(KIND_CLUSTER)
	@echo ""
	@echo "Local cluster ready. Run 'make deploy' to apply k8s manifests."

# Port-forward the divinity-server service to localhost:8080
port-forward:
	kubectl port-forward svc/divinity-server 8080:80 -n divinity

# Tear down the local Kind cluster
teardown-local:
	kind delete cluster --name $(KIND_CLUSTER)
