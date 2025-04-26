# GitHub Actions with AWS IAM Identity Federation

This document explains how to set up GitHub Actions with AWS IAM Identity Federation for secure, token-based authentication without long-lived credentials.

## Overview

AWS IAM Identity Federation allows GitHub Actions workflows to authenticate with AWS using short-lived credentials through OpenID Connect (OIDC). This approach eliminates the need to store long-lived AWS credentials as GitHub secrets.

## Benefits

- **Enhanced Security**: No long-lived access keys stored in GitHub secrets
- **Simplified Management**: No need to rotate credentials
- **Fine-grained Access Control**: Specific permissions for different environments
- **Conditional Access**: Limit access based on repository, branch, or other conditions

## Setup Instructions

### 1. Create an IAM OIDC Identity Provider for GitHub

```bash
# Create the OIDC provider for GitHub Actions
aws iam create-open-id-connect-provider \
  --url https://token.actions.githubusercontent.com \
  --client-id-list sts.amazonaws.com \
  --thumbprint-list 6938fd4d98bab03faadb97b34396831e3780aea1
```

### 2. Create IAM Roles for Different Environments

#### Create Trust Policy Files

Create a file named `trust-policy-preview-staging.json`:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::<ACCOUNT_ID>:oidc-provider/token.actions.githubusercontent.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
        },
        "StringLike": {
          "token.actions.githubusercontent.com:sub": "repo:bernhardrode/aws-infra-sandbox:*"
        }
      }
    }
  ]
}
```

Create a file named `trust-policy-production.json` with stricter conditions:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::<ACCOUNT_ID>:oidc-provider/token.actions.githubusercontent.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "token.actions.githubusercontent.com:aud": "sts.amazonaws.com",
          "token.actions.githubusercontent.com:sub": "repo:bernhardrode/aws-infra-sandbox:ref:refs/tags/*"
        }
      }
    }
  ]
}
```

#### Create IAM Roles

```bash
# Create role for preview/staging environments
aws iam create-role \
  --role-name GitHubActionsPreviewStaging \
  --assume-role-policy-document file://trust-policy-preview-staging.json

# Create role for production environment
aws iam create-role \
  --role-name GitHubActionsProduction \
  --assume-role-policy-document file://trust-policy-production.json
```

### 3. Attach Policies to the Roles

```bash
# Attach policies to preview/staging role
aws iam attach-role-policy \
  --role-name GitHubActionsPreviewStaging \
  --policy-arn arn:aws:iam::aws:policy/AmazonS3FullAccess

aws iam attach-role-policy \
  --role-name GitHubActionsPreviewStaging \
  --policy-arn arn:aws:iam::aws:policy/AmazonDynamoDBFullAccess

aws iam attach-role-policy \
  --role-name GitHubActionsPreviewStaging \
  --policy-arn arn:aws:iam::aws:policy/AmazonAPIGatewayAdministrator

aws iam attach-role-policy \
  --role-name GitHubActionsPreviewStaging \
  --policy-arn arn:aws:iam::aws:policy/AWSLambda_FullAccess

aws iam attach-role-policy \
  --role-name GitHubActionsPreviewStaging \
  --policy-arn arn:aws:iam::aws:policy/CloudWatchLogsFullAccess

# Attach policies to production role (same policies)
aws iam attach-role-policy \
  --role-name GitHubActionsProduction \
  --policy-arn arn:aws:iam::aws:policy/AmazonS3FullAccess

aws iam attach-role-policy \
  --role-name GitHubActionsProduction \
  --policy-arn arn:aws:iam::aws:policy/AmazonDynamoDBFullAccess

aws iam attach-role-policy \
  --role-name GitHubActionsProduction \
  --policy-arn arn:aws:iam::aws:policy/AmazonAPIGatewayAdministrator

aws iam attach-role-policy \
  --role-name GitHubActionsProduction \
  --policy-arn arn:aws:iam::aws:policy/AWSLambda_FullAccess

aws iam attach-role-policy \
  --role-name GitHubActionsProduction \
  --policy-arn arn:aws:iam::aws:policy/CloudWatchLogsFullAccess
```

> **Note**: For production use, it's recommended to create custom IAM policies with least privilege permissions rather than using the AWS managed policies shown above.

### 4. Configure GitHub Repository Secrets

In your GitHub repository, add the following secrets:

1. Go to your repository → Settings → Secrets and variables → Actions
2. Add the following repository secrets:

- `AWS_ROLE_TO_ASSUME`: `arn:aws:iam::<ACCOUNT_ID>:role/GitHubActionsPreviewStaging`
- `AWS_ROLE_TO_ASSUME_PROD`: `arn:aws:iam::<ACCOUNT_ID>:role/GitHubActionsProduction`
- `AWS_REGION`: Your AWS region (e.g., `eu-central-1`)

### 5. GitHub Actions Workflow Configuration

Our GitHub Actions workflows are already configured to use these roles. The key configuration is in the AWS credentials setup step:

```yaml
- name: Configure AWS credentials
  uses: aws-actions/configure-aws-credentials@v4
  with:
    role-to-assume: ${{ secrets.AWS_ROLE_TO_ASSUME }}
    aws-region: ${{ secrets.AWS_REGION }}
```

For production deployments:

```yaml
- name: Configure AWS credentials
  uses: aws-actions/configure-aws-credentials@v4
  with:
    role-to-assume: ${{ secrets.AWS_ROLE_TO_ASSUME_PROD }}
    aws-region: ${{ secrets.AWS_REGION }}
```

## Security Considerations

1. **Least Privilege**: Customize the IAM policies to grant only the permissions needed for your workflows.
2. **Branch Protection**: Enable branch protection rules for your main branch and require status checks to pass before merging.
3. **Environment Protection Rules**: For production deployments, consider setting up environment protection rules in GitHub.
4. **Audit Logging**: Enable AWS CloudTrail to log all API calls made by the GitHub Actions roles.

## Troubleshooting

### Common Issues

1. **Access Denied Errors**:
   - Verify the trust policy is correctly configured
   - Check that the GitHub repository name matches exactly what's in the trust policy
   - Ensure the IAM role has the necessary permissions

2. **Token Exchange Failures**:
   - Verify the OIDC provider is correctly set up
   - Check the thumbprint is correct
   - Ensure the `id-token: write` permission is set in your workflow

3. **Role Not Found**:
   - Double-check the role ARN in your GitHub secrets
   - Ensure the role exists in the specified AWS account

## References

- [AWS IAM Roles Anywhere Documentation](https://docs.aws.amazon.com/rolesanywhere/latest/userguide/introduction.html)
- [GitHub Actions OIDC Integration](https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/configuring-openid-connect-in-amazon-web-services)
- [AWS Security Blog: Use IAM Roles to Connect GitHub Actions to AWS](https://aws.amazon.com/blogs/security/use-iam-roles-to-connect-github-actions-to-actions-in-aws/)
## Automated Setup

This project includes an automated setup script to configure GitHub Actions with AWS IAM Identity Federation. To use it:

```bash
make setup-github
```

This command will:

1. Create an IAM OIDC Identity Provider for GitHub Actions
2. Create IAM roles for preview/staging and production environments
3. Set up the necessary trust policies
4. Configure GitHub repository secrets

### Prerequisites for Automated Setup

- AWS CLI installed and configured with appropriate permissions
- GitHub CLI installed and authenticated (optional, but recommended)
- Git repository with a remote pointing to GitHub

If the GitHub CLI is not available, the script will provide instructions for manually setting up the GitHub repository secrets.

## CDK Bootstrap Setup

For CDK deployments to work properly with GitHub Actions, you need to bootstrap your AWS account with the appropriate permissions. This project includes a command to handle this:

```bash
make bootstrap-cdk
```

This command will:

1. Bootstrap CDK in your AWS account
2. Configure the necessary trust relationships for CDK roles
3. Add required permissions to the GitHub Actions roles for accessing CDK bootstrap resources

### Common CDK Bootstrap Issues

If you encounter errors like:

```
AccessDeniedException: User is not authorized to perform: ssm:GetParameter on resource: arn:aws:ssm:***:parameter/cdk-bootstrap/hnb659fds/version
```

Or:

```
current credentials could not be used to assume 'arn:aws:iam::***:role/cdk-hnb659fds-deploy-role-***'
```

Run the `bootstrap-cdk` command to fix these permission issues.
## Combined Setup (Recommended)

For the smoothest experience, use the combined setup command that handles both GitHub Actions configuration and CDK bootstrapping in one step:

```bash
make setup
```

This single command will:

1. Set up the OIDC provider for GitHub Actions
2. Create IAM roles with appropriate permissions
3. Bootstrap CDK in your AWS account
4. Configure GitHub repository secrets
5. Set up all necessary trust relationships and permissions

This is the recommended approach for new projects.
