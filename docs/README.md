# AWS Infrastructure Sandbox Documentation

Welcome to the documentation for the AWS Infrastructure Sandbox project. This documentation will help you understand how to work with this project, deploy infrastructure, and contribute to its development.

## Table of Contents

1. [Development Workflow](./development-workflow.md)
   - Local development setup
   - Building and testing
   - Working with Lambda functions
   - Using the Makefile commands

2. [CI/CD Workflow](./ci-cd-workflow.md)
   - GitHub Actions workflows
   - PR pr environments
   - Staging and production deployments
   - Conventional commits

3. [AWS CDK Environment Management](./aws-cdk-environments.md)
   - Environment types
   - Resource naming conventions
   - Stack naming
   - Resource tagging

4. [GitHub Actions with AWS IAM Identity Federation](./github-aws-federation.md)
   - Setting up OIDC for GitHub Actions
   - Creating IAM roles
   - Security considerations
   - Troubleshooting

5. [Manual Deployment](./manual-deployment.md)
   - Deploying specific versions
   - Targeting different environments
   - Best practices for manual deployments

## Getting Started

To get started with development:

1. Clone the repository
2. Review the [Development Workflow](./development-workflow.md) documentation
3. Set up your local environment
4. Create your development stack with `make dev-create`

## Contributing

When contributing to this project:

1. Create a new branch for your feature or fix
2. Use conventional commits for your commit messages
3. Open a PR against the main branch
4. Review the automatically deployed pr environment
5. Address any feedback from reviewers

## Architecture

This project uses:

- AWS CDK for infrastructure as code
- Go for both infrastructure definition and Lambda functions
- GitHub Actions for CI/CD
- AWS IAM Identity Federation for secure authentication

## Security

For security considerations and best practices, refer to the [GitHub Actions with AWS IAM Identity Federation](./github-aws-federation.md) documentation.
