# AWS Infrastructure Sandbox with GitHub Actions CI/CD

This project is set up with GitHub Actions workflows to automate the deployment process across different environments:

## CI/CD Workflow

### PR Preview Environments
- When a PR is opened against the `main` branch, a preview environment is automatically deployed
- Each PR gets its own isolated environment with unique resource names
- A comment is added to the PR with links to the deployed resources
- Preview environments are automatically destroyed when the PR is closed

### Staging Environment
- When code is merged to `main`, it's automatically deployed to the staging environment
- Staging serves as a pre-release environment for testing before production

### Production Releases
- Uses the [Release Please](https://github.com/googleapis/release-please) action to automate versioning and releases
- When a release is created, the code is automatically deployed to production
- Follows semantic versioning based on conventional commits

## Environment Management

The project uses a custom environment management system to handle different deployment targets:

- **Development**: Local development environment
- **Preview**: Temporary environments for pull requests
- **Staging**: Pre-release environment for testing
- **Production**: Live environment for end users

## GitHub Actions Workflows

1. **PR Preview** (`pr-preview.yml`): Deploys preview environments for PRs
2. **Staging** (`staging.yml`): Deploys to staging when code is merged to main
3. **Release** (`release.yml`): Manages releases and deploys to production
4. **Test** (`test.yml`): Runs tests and CDK synthesis for validation

## Required Secrets

To use these workflows, you need to set up the following GitHub secrets:

- `AWS_ROLE_TO_ASSUME`: IAM role ARN for staging/preview environments
- `AWS_ROLE_TO_ASSUME_PROD`: IAM role ARN for production environment
- `AWS_REGION`: AWS region for deployments

## Conventional Commits

This project uses conventional commits to automate versioning:

- `feat:` - New features (minor version bump)
- `fix:` - Bug fixes (patch version bump)
- `chore:` - Maintenance tasks (no version bump)
- `docs:` - Documentation changes (no version bump)
- `perf:` - Performance improvements (patch version bump)
- `refactor:` - Code refactoring (no version bump)

Example: `feat: add new lambda function for user authentication`
