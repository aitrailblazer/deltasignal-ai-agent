APP_NAME ?= deltasignal-ai-agent
SERVICE ?= deltasignal-ai-agent
PROJECT ?= startup-ai-deltasignal
REGION ?= us-central1
IMAGE ?= $(REGION)-docker.pkg.dev/$(PROJECT)/$(APP_NAME)/app:local
PORT ?= 8080

-include .env
export

.PHONY: help test coverage vet check run demo spend build docker-build deploy deploy-source clean print-config

help:
	@echo "Available commands:"
	@echo "  make test          - Run Go tests"
	@echo "  make coverage      - Run Go tests with 100% coverage gate"
	@echo "  make vet           - Run go vet"
	@echo "  make check         - Run tests and vet"
	@echo "  make run           - Run the Cloud Run HTTP service locally"
	@echo "  make demo          - Run the two-turn TripCode demo client"
	@echo "  make spend         - Report project billing link and app usage ledger"
	@echo "  make build         - Build local server binary"
	@echo "  make docker-build  - Build local Docker image"
	@echo "  make deploy        - Deploy through cloudbuild.yaml"
	@echo "  make deploy-source - Deploy source directly to Cloud Run"
	@echo "  make clean         - Remove local build artifacts"

test:
	GOWORK=off go test ./...

coverage:
	GOWORK=off go test ./... -coverprofile=coverage.out -covermode=count
	GOWORK=off go tool cover -func=coverage.out
	@test "$$(GOWORK=off go tool cover -func=coverage.out | awk '/^total:/ {print $$3}')" = "100.0%" || (echo "coverage gate failed: total coverage is not 100.0%" && exit 1)

vet:
	GOWORK=off go vet ./...

check: test coverage vet

run:
	PORT=$(PORT) GOWORK=off go run ./cmd/server

demo:
	GOWORK=off go run ./cmd/tripcode-demo

spend:
	scripts/report-spend.sh

build:
	mkdir -p bin
	CGO_ENABLED=0 GOWORK=off go build -trimpath -ldflags="-w -s" -o bin/$(APP_NAME) ./cmd/server
	CGO_ENABLED=0 GOWORK=off go build -trimpath -ldflags="-w -s" -o bin/tripcode-demo ./cmd/tripcode-demo

docker-build:
	docker build -t $(APP_NAME):latest .

deploy:
	gcloud builds submit \
		--project=$(PROJECT) \
		--config=cloudbuild.yaml \
		--substitutions=_REGION=$(REGION),_IMAGE=$(IMAGE)

deploy-source:
	gcloud run deploy $(SERVICE) \
		--source . \
		--project=$(PROJECT) \
		--region=$(REGION) \
		--platform=managed \
		--allow-unauthenticated \
		--min-instances=0 \
		--max-instances=3 \
		--memory=512Mi \
		--cpu=1 \
		--concurrency=80 \
		--set-env-vars=GOOGLE_CLOUD_PROJECT=$(PROJECT),GOOGLE_CLOUD_LOCATION=global,GOOGLE_GENAI_USE_VERTEXAI=true,DELTASIGNAL_USE_GEMINI=true,GEMINI_MODEL=gemini-2.5-flash,DELTASIGNAL_USE_LIVE_MCP=true,DELTASIGNAL_ENABLE_TRIPCODE_FALLBACK=true,DELTASIGNAL_COST_TRACKING=true,DELTASIGNAL_GOOGLE_CREDIT_BUDGET_USD=500,DELTASIGNAL_ESTIMATED_BRIEF_COST_USD=0.02,DELTASIGNAL_ESTIMATED_TRIPCODE_COST_USD=0.05,DELTASIGNAL_ESTIMATED_SESSION_MEMORY_COST_USD=0.00,DELTASIGNAL_COST_SOURCE=local-estimate,DELTASIGNAL_BILLING_PROJECT_ID=$(PROJECT),DELTASIGNAL_BILLING_ENABLED=true,DELTASIGNAL_BILLING_ACCOUNT_NAME=billingAccounts/018530-615AB8-696753,DELTASIGNAL_RATE_LIMIT_ENABLED=true,DELTASIGNAL_RATE_LIMIT_PER_WINDOW=30,DELTASIGNAL_RATE_LIMIT_WINDOW_SECONDS=60,DELTASIGNAL_MEMORY_MAX_ENTRIES=20 \
		--set-secrets=MCP_API_KEY=deltasignal-mcp-api-key:latest,DELTASIGNAL_DEMO_API_KEY=deltasignal-demo-api-key:latest

print-config:
	@echo "APP_NAME=$(APP_NAME)"
	@echo "SERVICE=$(SERVICE)"
	@echo "PROJECT=$(PROJECT)"
	@echo "REGION=$(REGION)"
	@echo "IMAGE=$(IMAGE)"
	@echo "PORT=$(PORT)"

clean:
	rm -rf bin/
	docker rmi $(APP_NAME):latest 2>/dev/null || true
