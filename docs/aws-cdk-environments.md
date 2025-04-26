# AWS CDK Environment Management

This document explains how our project manages different deployment environments using AWS CDK.

## Environment Types

Our infrastructure supports four types of environments:

1. **Development**: Personal development environments for individual developers
2. **PR**: Temporary environments created for pull requests
3. **Staging**: Pre-production environment for testing
4. **Production**: Live environment for end users

## Environment Configuration

Environments are configured using CDK context values. These can be provided via the command line or in `cdk.json`.

### Context Parameters

- `environment`: The environment type (`dev`, `pr`, `staging`, `production`)
- `username`: The developer's username (for development environments)
- `prNumber`: The pull request number (for pr environments)
- `version`: The release version (for production environments)

## Environment-Specific Resource Naming

Resources are named differently based on the environment to avoid conflicts:

- **Development**: `{resourceName}-{username}-dev` (e.g., `api-ebbo-dev`)
- **PR**: `{resourceName}-pr{prNumber}` (e.g., `api-pr123`)
- **Staging**: `{resourceName}-staging` (e.g., `api-staging`)
- **Production**: `{resourceName}` (e.g., `api`)

## Stack Naming

Stack names follow a similar pattern:

- **Development**: `{username}-dev` (e.g., `ebbo-dev`)
- **PR**: `{stackName}-pr-{prNumber}` (e.g., `AwsInfraSandboxStack-pr-123`)
- **Staging**: `{stackName}-staging` (e.g., `AwsInfraSandboxStack-staging`)
- **Production**: `{stackName}` (e.g., `AwsInfraSandboxStack`)

## Resource Tagging

All resources are automatically tagged with:

- `Environment`: The environment type
- `ManagedBy`: Set to "CDK"
- `Owner`: The developer's username (for development environments)
- `PR`: Set to "true" (for pr environments)
- `PR`: The PR number (for pr environments)
- `Version`: The release version (for production environments)

## Usage Examples

### Deploying to Development Environment

```bash
make dev-deploy
```

This creates a stack named `ebbo-dev` with resources prefixed with your username.

### Deploying to PR Environment

```bash
cdk deploy --context environment=pr --context prNumber=123
```

This creates a stack named `AwsInfraSandboxStack-pr-123` with resources prefixed with `pr123`.

### Deploying to Staging Environment

```bash
cdk deploy --context environment=staging
```

This creates a stack named `AwsInfraSandboxStack-staging` with resources suffixed with `-staging`.

### Deploying to Production Environment

```bash
cdk deploy --context environment=production --context version=v1.0.0
```

This creates a stack named `AwsInfraSandboxStack` with resources using their base names.

## Implementation Details

The environment management is implemented in the `lib/environment.go` file, which provides:

- `Environment` struct to hold environment information
- `GetEnvironmentFromContext` function to extract environment info from CDK context
- `GetStackName` method to generate environment-specific stack names
- `GetResourceName` method to generate environment-specific resource names
- `GetTags` method to generate environment-specific resource tags

## Best Practices

1. **Always specify an environment**: Use the appropriate context parameters for your target environment
2. **Use the Makefile commands**: The Makefile provides shortcuts for common operations
3. **Clean up development environments**: Run `make dev-destroy` when you're done with your development environment
4. **Tag all resources**: Use the `environment.GetTags()` method to tag all resources
