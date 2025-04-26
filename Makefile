# Directory structure
CDK_DIR = infra
FUNCTIONS_DIR = functions
BUILD_DIR = build
BIN_DIR = $(shell pwd)/$(BUILD_DIR)/bin
DIST_DIR = $(shell pwd)/$(BUILD_DIR)/dist

# Go build flags
GOOS = linux
GOARCH = arm64
CGO_ENABLED = 0

# AWS CDK commands
CDK = cdk
CDK_APP = $(shell pwd)/$(CDK_DIR)/aws-infra-sandbox.go
CDK_BIN = $(BIN_DIR)/aws-infra-sandbox

# Get the current username from the environment
USERNAME := $(shell whoami)

# Environment settings
DEV_STACK_NAME = $(USERNAME)-dev

# Get function names
FUNCTION_NAMES = $(notdir $(wildcard $(FUNCTIONS_DIR)/*))

.PHONY: all clean build deploy destroy dev-deploy dev-destroy dev-diff watch-dev watch-dev-poll create update dev-create dev-update watch lambdas cdk-synth cdk-diff $(FUNCTION_NAMES)

# Default target
all: clean build

# Create build directories
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

$(BIN_DIR): 
	mkdir -p $(BIN_DIR)

$(DIST_DIR):
	mkdir -p $(DIST_DIR)

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -rf cdk.out
	rm -f infra/aws-infra-sandbox
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
		GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=$(CGO_ENABLED) go build -o $(BIN_DIR)/$$func && \
		cd $(BIN_DIR) && \
		cp $$func bootstrap && \
		chmod 755 bootstrap && \
		zip -j $(DIST_DIR)/$$func.zip bootstrap && \
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
	go build -o $(CDK_BIN)
	chmod +x $(CDK_BIN)
	@echo "CDK app built: $(CDK_BIN)"

# Create a new stack (alias for deploy)
create: deploy

# Update an existing stack (alias for deploy)
update: deploy

# CDK commands
cdk-synth:
	@echo "Synthesizing CDK stack..."
	$(CDK) synth --app $(CDK_BIN)

cdk-diff:
	@echo "Showing CDK diff..."
	$(CDK) diff --app $(CDK_BIN)

# Deploy the stack with the username in the stack name
deploy: build
	@echo "Deploying stack..."
	cd $(CDK_DIR) && $(CDK) deploy --app $(CDK_BIN) --all \
		--require-approval never \
		--context environment=development \
		--context username=$(USERNAME)

# Destroy the stack
destroy:
	@echo "Destroying stack..."
	$(CDK) destroy --app $(CDK_BIN) --force

# Development environment commands
dev-deploy: build
	@echo "Deploying development stack for $(USERNAME)..."
	cd $(CDK_DIR) && $(CDK) deploy --app $(CDK_BIN) --all \
		--require-approval never \
		--context environment=development \
		--context username=$(USERNAME)

dev-create: dev-deploy

dev-update: dev-deploy

dev-destroy:
	@echo "Destroying development stack for $(USERNAME)..."
	cd $(CDK_DIR) && $(CDK) destroy --app $(CDK_BIN) \
		--force \
		--require-approval never \
		--context environment=development \
		--context username=$(USERNAME)

dev-diff: build
	@echo "Showing diff for development stack..."
	cd $(CDK_DIR) && $(CDK) diff --app $(CDK_BIN) \
		--require-approval never \
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

# Help target
help:
	@echo "Available targets:"
	@echo "  all            - Clean, build, and deploy (default)"
	@echo "  clean          - Remove build artifacts"
	@echo "  build          - Build all Lambda functions and CDK app"
	@echo "  lambdas        - Build all Lambda functions"
	@echo "  create         - Create a new stack (alias for deploy)"
	@echo "  update         - Update an existing stack (alias for deploy)"
	@echo "  deploy         - Deploy the stack to AWS"
	@echo "  destroy        - Destroy the stack from AWS"
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
# Setup GitHub Actions with AWS IAM Identity Federation
setup-github:
	@echo "Setting up GitHub Actions with AWS IAM Identity Federation..."
	@./scripts/setup-github-aws-federation.sh
	@echo "  setup-github    - Set up GitHub Actions with AWS IAM Identity Federation"
