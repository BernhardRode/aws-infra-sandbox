# GitHub Actions with AWS IAM Identity Federation

This project uses GitHub Actions with AWS IAM Identity Federation for secure, credential-free deployments to AWS.

## Overview

Instead of storing long-lived AWS credentials as GitHub secrets, we use AWS IAM Identity Federation with OpenID Connect (OIDC). This approach:

1. Eliminates the need for long-lived AWS access keys
2. Provides fine-grained control over permissions
3. Improves security by using short-lived credentials
4. Allows different permission levels for different environments

## How It Works

1. GitHub Actions authenticates to AWS using OIDC
2. AWS verifies the OIDC token from GitHub
3. AWS grants temporary credentials based on the IAM role's permissions
4. GitHub Actions uses these temporary credentials to deploy resources

## IAM Roles

We use three separate IAM roles for different environments:

1. **GitHubActionsDevelopment**: Used for development and PR environments
   - Trust policy allows any workflow in the repository
   - Used for PR preview environments and local development

2. **GitHubActionsStaging**: Used for staging environment
   - Trust policy allows workflows from the main branch
   - Used for deployments to the staging environment

3. **GitHubActionsProduction**: Used for production environment
   - Trust policy only allows workflows triggered by tags
   - Used for production releases

## GitHub Secrets

The following GitHub secrets are set up:

- `AWS_ROLE_TO_ASSUME_DEVELOPMENT`: ARN of the development IAM role
- `AWS_ROLE_TO_ASSUME_STAGING`: ARN of the staging IAM role
- `AWS_ROLE_TO_ASSUME_PRODUCTION`: ARN of the production IAM role
- `AWS_REGION`: AWS region for deployments

For backward compatibility, we also set:
- `AWS_ROLE_TO_ASSUME`: Same as `AWS_ROLE_TO_ASSUME_DEVELOPMENT`
- `AWS_ROLE_TO_ASSUME_PROD`: Same as `AWS_ROLE_TO_ASSUME_PRODUCTION`

## Setup

The setup process is automated with the `setup-github-aws-federation.sh` script:

```bash
./scripts/setup-github-aws-federation.sh
```

This script:

1. Creates or updates the OIDC provider for GitHub Actions
2. Creates IAM roles with appropriate trust policies
3. Attaches necessary permissions to the roles
4. Sets up GitHub repository secrets

## Trust Policies

### Development Trust Policy

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::ACCOUNT_ID:oidc-provider/token.actions.githubusercontent.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
        },
        "StringLike": {
          "token.actions.githubusercontent.com:sub": "repo:OWNER/REPO:*"
        }
      }
    }
  ]
}
```

### Staging Trust Policy

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::ACCOUNT_ID:oidc-provider/token.actions.githubusercontent.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
        },
        "StringLike": {
          "token.actions.githubusercontent.com:sub": [
            "repo:OWNER/REPO:ref:refs/heads/main",
            "repo:OWNER/REPO:pull_request"
          ]
        }
      }
    }
  ]
}
```

### Production Trust Policy

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::ACCOUNT_ID:oidc-provider/token.actions.githubusercontent.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
        },
        "StringLike": {
          "token.actions.githubusercontent.com:sub": "repo:OWNER/REPO:ref:refs/tags/*"
        }
      }
    }
  ]
}
```

## Usage in Workflows

To use these roles in GitHub Actions workflows:

```yaml
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: ${{ secrets.AWS_ROLE_TO_ASSUME_DEVELOPMENT }}
          aws-region: ${{ secrets.AWS_REGION }}
```

For staging:

```yaml
role-to-assume: ${{ secrets.AWS_ROLE_TO_ASSUME_STAGING }}
```

For production:

```yaml
role-to-assume: ${{ secrets.AWS_ROLE_TO_ASSUME_PRODUCTION }}
```

## Troubleshooting

If you encounter issues with the OIDC authentication:

1. Check that the OIDC provider is correctly set up in AWS IAM
2. Verify that the trust policies are correctly configured
3. Ensure that the GitHub repository name matches the one in the trust policy
4. Check that the workflow is running from the expected branch or tag
5. Verify that the IAM roles have the necessary permissions

For more information, see the [AWS documentation on OIDC federation](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers_create_oidc.html).
