#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if AWS CLI is installed and configured
if ! command -v aws &> /dev/null; then
  echo -e "${RED}AWS CLI is not installed. Please install it first.${NC}"
  exit 1
fi

if ! aws sts get-caller-identity &> /dev/null; then
  echo -e "${RED}AWS CLI is not configured properly. Please run 'aws configure'.${NC}"
  exit 1
fi

# Check if CDK is installed
if ! command -v cdk &> /dev/null; then
  echo -e "${RED}AWS CDK is not installed. Please install it first with 'npm install -g aws-cdk'.${NC}"
  exit 1
fi

# Get AWS account ID and region
AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query "Account" --output text)
AWS_REGION=$(aws configure get region)
if [ -z "$AWS_REGION" ]; then
  echo -e "${YELLOW}AWS region not found in config. Using default: us-east-1${NC}"
  AWS_REGION="us-east-1"
fi

echo -e "${BLUE}Bootstrapping CDK in account ${AWS_ACCOUNT_ID} and region ${AWS_REGION}...${NC}"

# Bootstrap CDK with permissions for GitHub Actions roles
cdk bootstrap aws://${AWS_ACCOUNT_ID}/${AWS_REGION} \
  --cloudformation-execution-policies arn:aws:iam::aws:policy/AdministratorAccess \
  --trust ${AWS_ACCOUNT_ID} \
  --trust-for-lookup ${AWS_ACCOUNT_ID}

echo -e "${GREEN}CDK bootstrap completed successfully.${NC}"

# Update IAM roles with CDK bootstrap permissions if they exist
if aws iam get-role --role-name GitHubActionsStaging &> /dev/null; then
  echo -e "${BLUE}Adding CDK bootstrap permissions to GitHubActionsStaging role...${NC}"
  
  # Create inline policy for CDK bootstrap permissions
  cat > .aws-github-oidc/cdk-bootstrap-policy.json << EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "cloudformation:DescribeStacks",
        "ssm:GetParameter"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": "sts:AssumeRole",
      "Resource": [
        "arn:aws:iam::${AWS_ACCOUNT_ID}:role/cdk-*"
      ]
    }
  ]
}
EOF
  
  # Attach the policy to both roles
  aws iam put-role-policy \
    --role-name GitHubActionsStaging \
    --policy-name CDKBootstrapAccess \
    --policy-document file://.aws-github-oidc/cdk-bootstrap-policy.json
  
  if aws iam get-role --role-name GitHubActionsProduction &> /dev/null; then
    aws iam put-role-policy \
      --role-name GitHubActionsProduction \
      --policy-name CDKBootstrapAccess \
      --policy-document file://.aws-github-oidc/cdk-bootstrap-policy.json
  fi
  
  echo -e "${GREEN}CDK bootstrap permissions added to GitHub Actions roles.${NC}"
fi

echo -e "${GREEN}Setup complete! Your environment is now ready for CDK deployments.${NC}"
