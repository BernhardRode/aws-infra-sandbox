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
