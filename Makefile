# Directory structure
CDK_DIR = infra
FUNCTIONS_DIR = functions
BUILD_DIR = build
BIN_DIR = $(BUILD_DIR)/bin
DIST_DIR = $(BUILD_DIR)/dist

# Go build flags
GOOS = linux
GOARCH = arm64
CGO_ENABLED = 0

# AWS CDK commands
CDK = cdk
CDK_APP = $(shell pwd)/$(CDK_DIR)/aws-infra-sandbox.go

# Get function names
FUNCTION_NAMES = $(notdir $(wildcard $(FUNCTIONS_DIR)/*))

.PHONY: all clean build deploy destroy lambdas cdk-synth cdk-diff $(FUNCTION_NAMES)

# Default target
all: clean build deploy

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

# CDK commands
cdk-synth:
	@echo "Synthesizing CDK stack..."
	$(CDK) synth --app "go run $(CDK_APP)"

cdk-diff:
	@echo "Showing CDK diff..."
	$(CDK) diff --app "go run $(CDK_APP)"

# Deploy the stack
deploy: build
	@echo "Deploying stack..."
	@echo "Running CDK deploy with app: $(CDK_APP)"
	cd $(CDK_DIR) && $(CDK) deploy --app "go run aws-infra-sandbox.go" --require-approval never

# Destroy the stack
destroy:
	@echo "Destroying stack..."
	$(CDK) destroy --app "go run $(CDK_APP)" --force

# Run tests
test:
	@echo "Running tests..."
	@for dir in $(FUNCTION_DIRS); do \
		echo "Testing $$dir..."; \
		(cd $$dir && go test -v ./...); \
	done
	cd $(CDK_DIR) && go test -v ./...

# Development helper to watch for changes and rebuild
watch:
	@echo "Watching for changes..."
	find . -name "*.go" | entr -r make build

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
	@echo "  deploy         - Deploy the stack to AWS"
	@echo "  destroy        - Destroy the stack from AWS"
	@echo "  test           - Run tests"
	@echo "  cdk-synth      - Synthesize CDK stack"
	@echo "  cdk-diff       - Show changes to be deployed"
	@echo "  watch          - Watch for changes and rebuild"
	@echo "  list-functions - List all available functions"
	@echo "  debug          - Show debug information about paths"
