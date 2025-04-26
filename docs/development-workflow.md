# Development Workflow

This document explains the local development workflow for this project.

## Prerequisites

- AWS CLI configured with appropriate credentials
- Go 1.21 or later
- AWS CDK CLI installed (`npm install -g aws-cdk`)
- Optional: inotify-tools for efficient file watching (Linux only)

## Getting Started

1. Clone the repository:
   ```bash
   git clone https://github.com/bernhardrode/aws-infra-sandbox.git
   cd aws-infra-sandbox
   ```

2. Install dependencies:
   ```bash
   cd infra && go mod tidy && cd ..
   ```

## Development Commands

Our project includes a comprehensive Makefile with commands for common development tasks.

### Building the Project

```bash
make build
```

This command:
- Builds all Lambda functions in the `functions` directory
- Packages them as ZIP files in the `build/dist` directory
- Builds the CDK application

### Creating Your Development Stack

```bash
make dev-create
```

This command:
- Builds the project
- Deploys a development stack with your username
- Creates resources with names that include your username

### Updating Your Development Stack

```bash
make dev-update
```

This command updates your existing development stack with any changes.

### Watching for Changes

```bash
make watch
```

This command:
- Deploys your development stack
- Watches for changes in your code
- Automatically rebuilds and redeploys when changes are detected
- Uses the most efficient file watching method available on your system

### Viewing Deployment Changes

```bash
make dev-diff
```

This command shows what changes would be applied to your development stack.

### Destroying Your Development Stack

```bash
make dev-destroy
```

This command destroys your development stack and all associated resources.

## Working with Preview Environments

For pull request preview environments, you can use:

```bash
# Deploy a preview environment for a specific PR
make preview-deploy PR_NUMBER=123

# Destroy a preview environment for a specific PR
make preview-destroy PR_NUMBER=123
```

## Deploying to Different Environments

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

## Working with Lambda Functions

### Creating a New Lambda Function

1. Create a new directory in the `functions` directory:
   ```bash
   mkdir -p functions/my-new-function
   ```

2. Create a Go file with your Lambda code:
   ```bash
   touch functions/my-new-function/main.go
   ```

3. Implement your Lambda function:
   ```go
   package main

   import (
       "context"
       "github.com/aws/aws-lambda-go/lambda"
   )

   type Event struct {
       Name string `json:"name"`
   }

   type Response struct {
       Message string `json:"message"`
   }

   func HandleRequest(ctx context.Context, event Event) (Response, error) {
       return Response{
           Message: "Hello, " + event.Name,
       }, nil
   }

   func main() {
       lambda.Start(HandleRequest)
   }
   ```

4. Initialize the Go module:
   ```bash
   cd functions/my-new-function
   go mod init functions/my-new-function
   go mod tidy
   ```

5. Deploy your development stack:
   ```bash
   make dev-deploy
   ```

### Testing Lambda Functions

Run tests for all Lambda functions:

```bash
make test
```

## Git Workflow

1. Create a new branch for your feature:
   ```bash
   git checkout -b feature/my-new-feature
   ```

2. Make your changes and commit using conventional commits:
   ```bash
   git commit -m "feat: add new authentication feature"
   ```

3. Push your branch and create a PR:
   ```bash
   git push -u origin feature/my-new-feature
   ```

4. A preview environment will be automatically deployed for your PR

5. After review and approval, merge your PR to main

6. Your changes will be automatically deployed to staging

7. When ready for production, a release will be created and deployed

## Troubleshooting

### Common Issues

1. **Build Failures**:
   - Check that Go modules are properly initialized
   - Ensure all dependencies are installed
   - Verify that your code compiles locally

2. **Deployment Failures**:
   - Check AWS credentials are properly configured
   - Verify that you have the necessary permissions
   - Look for error messages in the CDK output

3. **Watch Mode Not Working**:
   - On Linux, install inotify-tools: `sudo apt-get install inotify-tools`
   - On other platforms, the polling-based watch mode will be used automatically

### Getting Help

If you encounter issues:

1. Check the error messages in the console
2. Review the AWS CloudFormation console for stack events
3. Check CloudWatch Logs for Lambda function logs
4. Consult the AWS CDK documentation

## Setting Up GitHub Actions with AWS

To set up GitHub Actions with AWS IAM Identity Federation:

```bash
make setup-github
```

This command configures the necessary AWS resources and GitHub repository secrets for secure authentication between GitHub Actions and AWS. See [GitHub Actions with AWS IAM Identity Federation](./github-aws-federation.md) for more details.

## Bootstrapping CDK

Before deploying with CDK for the first time, you need to bootstrap your AWS account:

```bash
make bootstrap-cdk
```

This command sets up the necessary resources in your AWS account for CDK deployments and configures the appropriate permissions for GitHub Actions.

## One-Step Setup (Recommended)

For the easiest setup experience, use the combined setup command:

```bash
make setup
```

This single command sets up both GitHub Actions with AWS IAM Identity Federation and bootstraps CDK in your AWS account. It's the recommended approach for new projects.
