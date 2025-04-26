# CI/CD Workflow

This document explains the CI/CD workflow implemented in this project using GitHub Actions.

## Overview

Our CI/CD pipeline automates the deployment process across different environments:

1. **PR PR Environments**: Automatically deployed when a PR is opened
2. **Staging Environment**: Automatically deployed when code is merged to main
3. **Production Environment**: Automatically deployed when a release is created

## GitHub Actions Workflows

### PR PR Workflow

**File**: `.github/workflows/pr-pr.yml`

This workflow is triggered when a pull request is opened, synchronized, or reopened against the main branch.

**Key Steps**:
1. Checkout code
2. Set up Go environment
3. Configure AWS credentials using OIDC federation
4. Install CDK and dependencies
5. Deploy a pr environment with a unique name based on the PR number
6. Add a comment to the PR with links to the deployed resources
7. Clean up the pr environment when the PR is closed

**Usage**:
- Simply open a PR against the main branch
- The workflow will automatically deploy a pr environment
- The PR will receive a comment with links to the deployed resources
- When the PR is closed, the pr environment is automatically destroyed

### Staging Workflow

**File**: `.github/workflows/staging.yml`

This workflow is triggered when code is pushed to the main branch.

**Key Steps**:
1. Checkout code
2. Set up Go environment
3. Configure AWS credentials using OIDC federation
4. Install CDK and dependencies
5. Deploy to the staging environment

**Usage**:
- Merge a PR to the main branch
- The workflow will automatically deploy to the staging environment

### Release Workflow

**File**: `.github/workflows/release.yml`

This workflow is triggered when code is pushed to the main branch and uses Release Please to manage releases.

**Key Steps**:
1. Run Release Please to determine if a new release should be created
2. If a new release is created:
   - Checkout code
   - Set up Go environment
   - Configure AWS credentials using OIDC federation for production
   - Install CDK and dependencies
   - Deploy to the production environment with the release version

**Usage**:
- Use conventional commits (e.g., `feat:`, `fix:`) in your PRs
- When appropriate commits are merged to main, Release Please will create a release PR
- When the release PR is merged, a new release is created
- The workflow will automatically deploy to the production environment

### Test Workflow

**File**: `.github/workflows/test.yml`

This workflow is triggered on all pushes to the main branch and on all PRs.

**Key Steps**:
1. Checkout code
2. Set up Go environment
3. Install dependencies
4. Run tests
5. Synthesize CDK templates

**Usage**:
- This workflow runs automatically on all PRs and pushes to main
- It ensures that the code builds correctly and passes all tests

## Conventional Commits

This project uses conventional commits to automate versioning:

- `feat:` - New features (minor version bump)
- `fix:` - Bug fixes (patch version bump)
- `chore:` - Maintenance tasks (no version bump)
- `docs:` - Documentation changes (no version bump)
- `perf:` - Performance improvements (patch version bump)
- `refactor:` - Code refactoring (no version bump)

Example: `feat: add new lambda function for user authentication`

## Required GitHub Secrets

To use these workflows, you need to set up the following GitHub secrets:

- `AWS_ROLE_TO_ASSUME`: IAM role ARN for staging/pr environments
- `AWS_ROLE_TO_ASSUME_PROD`: IAM role ARN for production environment
- `AWS_REGION`: AWS region for deployments

## Workflow Diagram

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  Developer  │     │  PR PR │     │   Staging   │     │  Production │
│ Environment │     │ Environment │     │ Environment │     │ Environment │
└──────┬──────┘     └──────┬──────┘     └──────┬──────┘     └──────┬──────┘
       │                   │                   │                   │
       │                   │                   │                   │
┌──────▼──────┐     ┌──────▼──────┐     ┌──────▼──────┐     ┌──────▼──────┐
│  Local Dev  │     │  Open PR    │     │  Merge PR   │     │   Release   │
│             │────►│             │────►│  to Main    │────►│   Created   │
│ make watch  │     │ Auto-deploy │     │ Auto-deploy │     │ Auto-deploy │
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
```

## Best Practices

1. **Use conventional commits**: This helps with automatic versioning
2. **Review PR prs**: Always check the pr environment before merging
3. **Test in staging**: Verify changes in staging before creating a release
4. **Monitor deployments**: Watch the GitHub Actions logs for any deployment issues
## Manual Deployment

In addition to the automated CI/CD workflows, this project also includes a manual deployment workflow for controlled deployments of specific versions:

### Manual Deployment Workflow

**File**: `.github/workflows/manual-deploy.yml`

This workflow is triggered manually through the GitHub Actions UI and allows you to:

1. Deploy any Git reference (tag, branch, or commit) to either staging or production
2. Specify a custom version label for the deployment
3. Get a detailed deployment summary

**Key Features**:
- Choose between staging and production environments
- Deploy any Git reference (tag, branch, or commit)
- Optionally specify a custom version label
- Get a detailed deployment summary

**Usage**:
- Go to the "Actions" tab in your repository
- Select the "Manual Deployment" workflow
- Click "Run workflow"
- Fill in the parameters and click "Run workflow" again

For more details, see the [Manual Deployment](./manual-deployment.md) documentation.
