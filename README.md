# AWS Infrastructure Sandbox

A CDK Go project with GitHub Actions CI/CD for automated deployments across multiple environments.

## Features

- **Multi-environment Infrastructure**: Development, PR, Staging, and Production environments
- **GitHub Actions CI/CD**: Automated deployments for PRs, staging, and production
- **AWS IAM Identity Federation**: Secure authentication without long-lived credentials
- **Developer-friendly Workflow**: Easy local development with watch mode

## Getting Started

### Prerequisites

- AWS CLI configured with appropriate credentials
- Go 1.21 or later
- AWS CDK CLI installed (`npm install -g aws-cdk`)
- Optional (development only): inotify-tools for efficient file watching (Linux only/Not in CICD)

### Quick Start

1. Clone the repository:
   ```bash
   git clone https://github.com/bernhardrode/aws-infra-sandbox.git
   cd aws-infra-sandbox
   ```

2. Create your development stack:
   ```bash
   make dev-create
   ```

3. Start development with watch mode:
   ```bash
   make watch
   ```

## Development Workflow

- **Local Development**: `make dev-create` and `make watch`
- **View Changes**: `make dev-diff`
- **Clean Up**: `make dev-destroy`

See [Development Workflow](./docs/development-workflow.md) for more details.

## CI/CD Pipeline

- **PR Preview**: Automatically deployed when a PR is opened
- **Staging**: Automatically deployed when code is merged to main
- **Production**: Automatically deployed when a release is created

See [CI/CD Workflow](./docs/ci-cd-workflow.md) for more details.

## Project Structure

```
.
├── .github/workflows/    # GitHub Actions workflows
├── build/                # Build artifacts
├── docs/                 # Documentation
├── functions/            # Lambda function code
├── infra/                # CDK infrastructure code
│   └── lib/              # Shared infrastructure libraries
├── .release-please-config.json  # Release configuration
├── cdk.json              # CDK configuration
├── Makefile              # Build and deployment commands
└── README.md             # This file
```

## Available Commands

Run `make help` to see all available commands:

```
Available targets:
  all            - Clean, build, and deploy (default)
  clean          - Remove build artifacts
  build          - Build all Lambda functions and CDK app
  lambdas        - Build all Lambda functions
  create         - Create a new stack (alias for deploy)
  update         - Update an existing stack (alias for deploy)
  deploy         - Deploy the stack to AWS (use ENVIRONMENT=preview|staging|production)
  destroy        - Destroy the stack from AWS (use ENVIRONMENT=preview|staging|production)
  preview-deploy - Deploy preview environment (requires PR_NUMBER)
  preview-destroy - Destroy preview environment (requires PR_NUMBER)
  dev-create     - Create development stack for current user
  dev-update     - Update development stack for current user
  dev-deploy     - Deploy development stack for current user
  dev-destroy    - Destroy development stack for current user
  dev-diff       - Show changes to be deployed to development stack
  watch          - Watch for changes and auto-deploy (smart detection)
  watch-dev      - Watch for changes with inotify (requires inotify-tools)
  watch-dev-poll - Watch for changes using polling (no dependencies)
  test           - Run tests
  cdk-synth      - Synthesize CDK stack
  cdk-diff       - Show changes to be deployed
  list-functions - List all available functions
```

## Environment Management

The project supports multiple deployment environments:

- **Development**: Personal environments for individual developers
- **Preview**: Temporary environments for pull requests
- **Staging**: Pre-production environment for testing
- **Production**: Live environment for end users

You can deploy to different environments using:

```bash
# Development environment (personal)
make dev-deploy

# Preview environment (for PRs)
make preview-deploy PR_NUMBER=123

# Staging environment
make deploy ENVIRONMENT=staging

# Production environment
make deploy ENVIRONMENT=production VERSION=v1.0.0
```

See [AWS CDK Environment Management](./docs/aws-cdk-environments.md) for more details.

## Documentation

For more detailed documentation, see the [docs](./docs) directory:

- [Development Workflow](./docs/development-workflow.md)
- [CI/CD Workflow](./docs/ci-cd-workflow.md)
- [AWS CDK Environment Management](./docs/aws-cdk-environments.md)
- [GitHub Actions with AWS IAM Identity Federation](./docs/github-aws-federation.md)

## Contributing

1. Create a new branch for your feature or fix
2. Use conventional commits for your commit messages:
   - `feat:` - New features (minor version bump)
   - `fix:` - Bug fixes (patch version bump)
   - `chore:` - Maintenance tasks (no version bump)
   - `docs:` - Documentation changes (no version bump)
   - `perf:` - Performance improvements (patch version bump)
   - `refactor:` - Code refactoring (no version bump)
3. Open a PR against the main branch
4. Review the automatically deployed preview environment
5. Address any feedback from reviewers

## License

This project is licensed under the MIT License - see the LICENSE file for details.
## Vaultwarden on ECS Fargate

This project now includes a Vaultwarden password manager deployment on AWS ECS Fargate. The implementation:

- Creates a dedicated VPC with public and private subnets
- Sets up VPC endpoints for secure access to AWS services without internet gateways
- Deploys an ECS Fargate cluster with the Vaultwarden container
- Uses EFS for persistent storage of Vaultwarden data
- Provides an Application Load Balancer for public access
- Supports optional HTTPS with automatic certificate provisioning

### Configuration

You can configure the Vaultwarden deployment by modifying the `vaultwardenConfig` object in the main file (`infra/aws-infra-sandbox.go`):

```go
vaultwardenConfig := &vaultwarden.VaultwardenConfig{
    // Base configuration
    BaseImageName: "vaultwarden/server",
    BaseVersion:   "latest",
    DomainName:    "", // Set via VAULTWARDEN_DOMAIN_NAME env var or leave empty for HTTP
    
    // VPC configuration
    VpcCidr:  "20.0.0.0/24",
    MaxAzs:   2,
    
    // ECS configuration
    ClusterName:  "vaultwarden-cluster",
    DesiredCount: 1,
    Cpu:          256, // 0.25 vCPU
    MemoryMiB:    512, // 512 MB RAM
    
    // EFS configuration
    FileSystemName:           "vaultwarden-fs",
    EnableAutomaticBackups:   true,
    LifecyclePolicyDays:      14,
    OutOfInfrequentAccessHits: 1,
}
```

You can also override some settings using environment variables:

- `VAULTWARDEN_BASE_VERSION`: The version of the Vaultwarden image to use (defaults to "latest")
- `VAULTWARDEN_DOMAIN_NAME`: Optional domain name for HTTPS access
- `VAULTWARDEN_CONFIG_*`: Any environment variables prefixed with `VAULTWARDEN_CONFIG_` will be passed to the container

### Deployment

To deploy the Vaultwarden stack:

```bash
# Development environment
make dev-deploy

# Staging environment
make deploy ENVIRONMENT=staging

# Production environment
make deploy ENVIRONMENT=production
```

When providing a domain name, the deployment will create an SSL certificate and wait for DNS validation before completing.
