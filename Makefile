# Directory structure
CDK_DIR = infra
FUNCTIONS_DIR = functions
BUILD_DIR = build
BIN_DIR = $(BUILD_DIR)/bin
DIST_DIR = $(BUILD_DIR)/dist
CDK_OUT_DIR = $(BUILD_DIR)/cdk.out

# Go build flags
GOOS = linux
GOARCH = arm64
CGO_ENABLED = 0

# AWS CDK commands
CDK = cdk
CDK_APP = $(shell pwd)/$(CDK_DIR)/aws-infra-sandbox.go
CDK_BIN = "$(CURDIR)/$(BIN_DIR)/aws-infra-sandbox"
CDK_OUTDIR_OPTION = --output $(CDK_OUT_DIR)

# Default values for environment variables
USERNAME = $(shell whoami)
ENVIRONMENT ?= development
PR_NUMBER ?=
SHA ?= $(shell git rev-parse --short HEAD)

# Get function names
FUNCTION_NAMES = $(notdir $(wildcard $(FUNCTIONS_DIR)/*))

# Define targets for different environments
.PHONY: all clean build deploy destroy dev-deploy dev-destroy dev-diff watch-dev watch-dev-poll create update dev-create dev-update watch lambdas cdk-synth cdk-diff $(FUNCTION_NAMES) setup-github bootstrap-cdk setup pr-deploy pr-destroy preview-deploy preview-destroy

# Default target
all: clean build deploy

# Create build directories
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

$(BIN_DIR): 
	mkdir -p $(BIN_DIR)

$(DIST_DIR):
	mkdir -p $(DIST_DIR)

$(CDK_OUT_DIR):
	mkdir -p $(CDK_OUT_DIR)

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	@echo "Clean complete"

# Build all Lambda functions
lambdas: $(BIN_DIR) $(DIST_DIR)
	@echo "Building Lambda functions..."
	@for func in $(FUNCTION_NAMES); do \
		echo "Building Lambda function: $$func"; \
		cd $(FUNCTIONS_DIR)/$$func && \
		if [ ! -f "go.mod" ]; then \
			go mod init $(FUNCTIONS_DIR)/$$func; \
			go mod tidy; \
		fi && \
		GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=$(CGO_ENABLED) go build -o $(CURDIR)/$(BIN_DIR)/$$func && \
		cd $(CURDIR)/$(BIN_DIR) && \
		cp $$func bootstrap && \
		chmod 755 bootstrap && \
		zip -j $(CURDIR)/$(DIST_DIR)/$$func.zip bootstrap && \
		rm bootstrap && \
		echo "Lambda build complete: $$func"; \
	done
	@echo "All Lambda functions built"

# Build both Lambda and CDK
build: $(BIN_DIR) lambdas
	@echo "Building CDK app..."
	@echo "CDK_DIR: $(CDK_DIR)"
	@cd $(CDK_DIR) && \
	if [ ! -f "go.mod" ]; then \
		go mod init aws-infra-sandbox && \
		go mod tidy; \
	fi && \
	go build -o $(CURDIR)/$(BIN_DIR)/aws-infra-sandbox
	@echo "Build complete"

# Create a new stack (alias for deploy)
create: deploy

# Update an existing stack (alias for deploy)
update: deploy

# CDK commands
cdk-synth: build $(CDK_OUT_DIR)
	@echo "Synthesizing CDK stack..."
	$(CDK) synth --app $(CDK_BIN) $(CDK_OUTDIR_OPTION)

cdk-diff: build $(CDK_OUT_DIR)
	@echo "Showing CDK diff..."
	$(CDK) diff --app $(CDK_BIN) $(CDK_OUTDIR_OPTION)

# Deploy the stack
deploy: build $(CDK_OUT_DIR)
	@echo "Deploying stack for environment: $(ENVIRONMENT)..."
	@echo "Running CDK deploy with app: $(CDK_BIN)"
	$(CDK) deploy --app $(CDK_BIN) $(CDK_OUTDIR_OPTION) --all \
		--require-approval never \
		--context environment=$(ENVIRONMENT) \
		$(if $(PR_NUMBER),--context pr_number=$(PR_NUMBER),) \
		$(if $(USERNAME),--context username=$(USERNAME),) \
		$(if $(VERSION),--context version=$(VERSION),) \
		$(if $(SHA),--context sha=$(SHA),)

# Destroy the stack
destroy: build $(CDK_OUT_DIR)
	@echo "Destroying stack for environment: $(ENVIRONMENT)..."
	$(CDK) destroy --app $(CDK_BIN) $(CDK_OUTDIR_OPTION) --all \
		--force \
		--context environment=$(ENVIRONMENT) \
		$(if $(PR_NUMBER),--context pr_number=$(PR_NUMBER),) \
		$(if $(USERNAME),--context username=$(USERNAME),) \
		$(if $(SHA),--context sha=$(SHA),)

# Development environment commands
dev-deploy: build $(CDK_OUT_DIR)
	@echo "Deploying development stack for $(USERNAME)..."
	$(CDK) deploy --app $(CDK_BIN) $(CDK_OUTDIR_OPTION) --all \
		--require-approval never \
		--context environment=development \
		--context username=$(USERNAME) \
		$(if $(SHA),--context sha=$(SHA),)

# PR environment commands
preview-deploy: pr-deploy

preview-destroy: pr-destroy

pr-deploy:
	@$(MAKE) deploy ENVIRONMENT=pr $(if $(PR_NUMBER),,$(error PR_NUMBER is required for pr environment))

pr-destroy:
	@$(MAKE) destroy ENVIRONMENT=pr $(if $(PR_NUMBER),,$(error PR_NUMBER is required for pr environment))

dev-create: dev-deploy

dev-update: dev-deploy

dev-destroy: build $(CDK_OUT_DIR)
	@echo "Destroying development stack for $(USERNAME)..."
	$(CDK) destroy --app $(CDK_BIN) $(CDK_OUTDIR_OPTION) --all \
		--force \
		--context environment=development \
		--context username=$(USERNAME)

dev-diff: build $(CDK_OUT_DIR)
	@echo "Showing diff for development stack..."
	$(CDK) diff --app $(CDK_BIN) $(CDK_OUTDIR_OPTION) --all \
		--context environment=development \
		--context username=$(USERNAME)

# Watch mode for development
watch-dev:
	@echo "Watching for changes and deploying to development environment..."
	@echo "Press Ctrl+C to stop watching"
	@echo "Note: This requires inotifywait. Install with: sudo apt-get install inotify-tools"
	@if command -v inotifywait > /dev/null; then \
		touch .watch-timestamp; \
		while true; do \
			make dev-deploy; \
			echo "Watching for changes. Press Ctrl+C to stop."; \
			inotifywait -r -e modify,create,delete,move $(CDK_DIR) $(FUNCTIONS_DIR) || break; \
			echo "Changes detected, rebuilding and redeploying..."; \
		done; \
	else \
		echo "Error: inotifywait not found. Install with: sudo apt-get install inotify-tools"; \
		echo "Falling back to polling mode..."; \
		make watch-dev-poll; \
	fi

# Alternative watch mode using polling (no dependencies)
watch-dev-poll:
	@echo "Watching for changes and deploying to development environment (polling mode)..."
	@echo "Press Ctrl+C to stop watching"
	@touch .watch-timestamp; \
	while true; do \
		make dev-deploy; \
		echo "Watching for changes. Press Ctrl+C to stop."; \
		sleep 5; \
		if [ -n "$$(find $(CDK_DIR) $(FUNCTIONS_DIR) -type f -name "*.go" -newer .watch-timestamp 2>/dev/null)" ]; then \
			echo "Changes detected, rebuilding and redeploying..."; \
			touch .watch-timestamp; \
		fi; \
	done

# General watch command that chooses the appropriate implementation
watch:
	@if command -v inotifywait > /dev/null; then \
		make watch-dev; \
	else \
		make watch-dev-poll; \
	fi

# Run tests
test:
	@echo "Running tests..."
	@for dir in $(FUNCTION_DIRS); do \
		echo "Testing $$dir..."; \
		(cd $$dir && go test -v ./...); \
	done
	cd $(CDK_DIR) && go test -v ./...

# List all functions
list-functions:
	@echo "Available functions:"
	@for func in $(FUNCTION_NAMES); do \
		echo "  $$func"; \
	done

# Setup GitHub Actions with AWS IAM Identity Federation
setup-github:
	@echo "Setting up GitHub Actions with AWS IAM Identity Federation..."
	@./scripts/setup-github-aws-federation.sh

# Bootstrap CDK for deployments
bootstrap-cdk:
	@echo "Bootstrapping CDK in your AWS account..."
	@./scripts/bootstrap-cdk.sh

# Setup DNS records for ebbo.dev
setup-dns:
	@echo "Setting up DNS records for ebbo.dev..."
	@./scripts/setup-dns-records.sh

# Complete setup for GitHub Actions and CDK
setup:
	@echo "Setting up GitHub Actions with AWS IAM Identity Federation, CDK Bootstrap, and DNS records..."
	@./scripts/setup-github-aws-complete.sh
	@./scripts/setup-dns-records.sh

# Help target
help:
	@echo "Available targets:"
	@echo "  all            - Clean, build, and deploy (default)"
	@echo "  clean          - Remove build artifacts"
	@echo "  build          - Build all Lambda functions and CDK app"
	@echo "  lambdas        - Build all Lambda functions"
	@echo "  create         - Create a new stack (alias for deploy)"
	@echo "  update         - Update an existing stack (alias for deploy)"
	@echo "  deploy         - Deploy the stack to AWS (use ENVIRONMENT=pr|staging|production)"
	@echo "  destroy        - Destroy the stack from AWS (use ENVIRONMENT=pr|staging|production)"
	@echo "  pr-deploy - Deploy pr environment (requires PR_NUMBER)"
	@echo "  pr-destroy - Destroy pr environment (requires PR_NUMBER)"
	@echo "  dev-create     - Create development stack for $(USERNAME)"
	@echo "  dev-update     - Update development stack for $(USERNAME)"
	@echo "  dev-deploy     - Deploy development stack for $(USERNAME)"
	@echo "  dev-destroy    - Destroy development stack for $(USERNAME)"
	@echo "  dev-diff       - Show changes to be deployed to development stack"
	@echo "  watch          - Watch for changes and auto-deploy (smart detection)"
	@echo "  watch-dev      - Watch for changes with inotify (requires inotify-tools)"
	@echo "  watch-dev-poll - Watch for changes using polling (no dependencies)"
	@echo "  test           - Run tests"
	@echo "  cdk-synth      - Synthesize CDK stack"
	@echo "  cdk-diff       - Show changes to be deployed"
	@echo "  list-functions - List all available functions"
	@echo "  setup-github   - Set up GitHub Actions with AWS IAM Identity Federation"
	@echo "  bootstrap-cdk  - Bootstrap CDK in your AWS account"
	@echo "  setup-dns      - Set up DNS records for ebbo.dev domain"
	@echo "  setup          - Complete setup for GitHub Actions, CDK, and DNS (recommended)"
