#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Global variables
HAS_GH=false
AWS_ACCOUNT_ID=""
AWS_REGION=""
GITHUB_OWNER=""
GITHUB_REPO_NAME=""
DEVELOPMENT_ROLE_ARN=""
STAGING_ROLE_ARN=""
PRODUCTION_ROLE_ARN=""
POLICY_DIR=".aws-github-oidc"

# Check if required tools are installed
check_requirements() {
  echo -e "${BLUE}Checking requirements...${NC}"
  
  # Check AWS CLI
  if ! command -v aws &> /dev/null; then
    echo -e "${RED}AWS CLI is not installed. Please install it first.${NC}"
    echo "Visit: https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html"
    exit 1
  fi
  
  # Check GitHub CLI
  if ! command -v gh &> /dev/null; then
    echo -e "${YELLOW}GitHub CLI is not installed. Some features will be limited.${NC}"
    echo "For full functionality, install GitHub CLI: https://cli.github.com/manual/installation"
    HAS_GH=false
  else
    HAS_GH=true
  fi
  
  # Check if AWS CLI is configured
  if ! aws sts get-caller-identity &> /dev/null; then
    echo -e "${RED}AWS CLI is not configured properly. Please run 'aws configure'.${NC}"
    exit 1
  fi
  
  # Check if GitHub CLI is authenticated (if installed)
  if [ "$HAS_GH" = true ] && ! gh auth status &> /dev/null; then
    echo -e "${YELLOW}GitHub CLI is not authenticated. Please run 'gh auth login'.${NC}"
    HAS_GH=false
  fi
  
  echo -e "${GREEN}All required tools are available.${NC}"
}

# Get AWS account ID
get_aws_account_id() {
  AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query "Account" --output text)
  echo -e "${BLUE}Using AWS Account ID: ${AWS_ACCOUNT_ID}${NC}"
}

# Get AWS region
get_aws_region() {
  AWS_REGION=$(aws configure get region)
  if [ -z "$AWS_REGION" ]; then
    echo -e "${YELLOW}AWS region not found in config. Using default: us-east-1${NC}"
    AWS_REGION="us-east-1"
  fi
  echo -e "${BLUE}Using AWS Region: ${AWS_REGION}${NC}"
}

# Get GitHub repository information
get_github_repo_info() {
  # Try to get from git remote
  GITHUB_REPO_URL=$(git config --get remote.origin.url)
  
  if [ -z "$GITHUB_REPO_URL" ]; then
    echo -e "${YELLOW}Could not determine GitHub repository from git config.${NC}"
    read -p "Please enter your GitHub repository (format: owner/repo): " GITHUB_REPO
  else
    # Extract owner/repo from URL
    if [[ $GITHUB_REPO_URL == *"github.com"* ]]; then
      GITHUB_REPO=$(echo $GITHUB_REPO_URL | sed -n 's/.*github.com[:/]\([^.]*\).*/\1/p')
    else
      echo -e "${YELLOW}Could not parse GitHub repository from URL.${NC}"
      read -p "Please enter your GitHub repository (format: owner/repo): " GITHUB_REPO
    fi
  fi
  
  # Split into owner and repo
  GITHUB_OWNER=$(echo $GITHUB_REPO | cut -d '/' -f 1)
  GITHUB_REPO_NAME=$(echo $GITHUB_REPO | cut -d '/' -f 2)
  
  echo -e "${BLUE}Using GitHub repository: ${GITHUB_OWNER}/${GITHUB_REPO_NAME}${NC}"
}

# Create or update IAM OIDC provider for GitHub Actions
create_oidc_provider() {
  echo -e "${BLUE}Creating or updating IAM OIDC provider for GitHub Actions...${NC}"
  
  # Check if provider already exists
  if aws iam list-open-id-connect-providers | grep -q "token.actions.githubusercontent.com"; then
    echo -e "${BLUE}OIDC provider for GitHub Actions already exists. Checking configuration...${NC}"
    
    # Get the ARN of the existing provider
    PROVIDER_ARN=$(aws iam list-open-id-connect-providers | grep -o "arn:aws:iam::${AWS_ACCOUNT_ID}:oidc-provider/token.actions.githubusercontent.com")
    
    # Update the thumbprint list
    echo -e "${BLUE}Updating thumbprint list for OIDC provider...${NC}"
    aws iam update-open-id-connect-provider-thumbprint \
      --open-id-connect-provider-arn "$PROVIDER_ARN" \
      --thumbprint-list 6938fd4d98bab03faadb97b34396831e3780aea1
    
    echo -e "${GREEN}OIDC provider updated successfully.${NC}"
  else
    # Create the OIDC provider
    aws iam create-open-id-connect-provider \
      --url https://token.actions.githubusercontent.com \
      --client-id-list sts.amazonaws.com \
      --thumbprint-list 6938fd4d98bab03faadb97b34396831e3780aea1
    
    echo -e "${GREEN}OIDC provider created successfully.${NC}"
  fi
}

# Create a trust policy file
create_trust_policy() {
  local policy_name=$1
  local conditions=$2
  
  echo -e "${BLUE}Creating trust policy: ${policy_name}...${NC}"
  
  cat > "${POLICY_DIR}/${policy_name}.json" << EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::${AWS_ACCOUNT_ID}:oidc-provider/token.actions.githubusercontent.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
        },
        "StringLike": {
          "token.actions.githubusercontent.com:sub": ${conditions}
        }
      }
    }
  ]
}
EOF
}

# Create all trust policies
create_trust_policies() {
  echo -e "${BLUE}Creating trust policy files...${NC}"
  
  # Create directory for policies if it doesn't exist
  mkdir -p ${POLICY_DIR}
  
  # Development trust policy
  create_trust_policy "trust-policy-development" '[
    "repo:${GITHUB_OWNER}/${GITHUB_REPO_NAME}:pull_request",
    "repo:${GITHUB_OWNER}/${GITHUB_REPO_NAME}:ref:refs/heads/feature/*",
    "repo:${GITHUB_OWNER}/${GITHUB_REPO_NAME}:ref:refs/heads/bugfix/*",
    "repo:${GITHUB_OWNER}/${GITHUB_REPO_NAME}:ref:refs/heads/dev/*"
  ]'
  
  # Staging trust policy
  create_trust_policy "trust-policy-staging" '"repo:${GITHUB_OWNER}/${GITHUB_REPO_NAME}:ref:refs/heads/main"'
  
  # Production trust policy
  create_trust_policy "trust-policy-production" '"repo:${GITHUB_OWNER}/${GITHUB_REPO_NAME}:ref:refs/tags/v*"'

  # Create CDK bootstrap policy
  cat > ${POLICY_DIR}/cdk-bootstrap-policy.json << EOF
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
  
  echo -e "${GREEN}Trust policy files created in ${POLICY_DIR}/ directory.${NC}"
}

# Create a deployment policy
create_deployment_policy() {
  local role_name=$1
  
  echo -e "${BLUE}Creating deployment policy for ${role_name}...${NC}"
  
  cat > ${POLICY_DIR}/policy-${role_name}.json << EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "cloudformation:CreateStack",
        "cloudformation:DeleteStack",
        "cloudformation:DescribeStacks",
        "cloudformation:DescribeStackEvents",
        "cloudformation:DescribeStackResources",
        "cloudformation:GetTemplate",
        "cloudformation:ValidateTemplate",
        "cloudformation:UpdateStack",
        "cloudformation:ListStacks"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:CreateBucket",
        "s3:DeleteBucket",
        "s3:ListBucket",
        "s3:PutObject",
        "s3:GetObject",
        "s3:DeleteObject",
        "s3:PutBucketPolicy",
        "s3:GetBucketPolicy"
      ],
      "Resource": [
        "arn:aws:s3:::cdk-*",
        "arn:aws:s3:::cdk-*/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "lambda:CreateFunction",
        "lambda:DeleteFunction",
        "lambda:GetFunction",
        "lambda:UpdateFunctionCode",
        "lambda:UpdateFunctionConfiguration",
        "lambda:AddPermission",
        "lambda:RemovePermission",
        "lambda:InvokeFunction",
        "lambda:GetPolicy"
      ],
      "Resource": "arn:aws:lambda:${AWS_REGION}:${AWS_ACCOUNT_ID}:function:*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "apigateway:GET",
        "apigateway:POST",
        "apigateway:PUT",
        "apigateway:DELETE",
        "apigateway:PATCH"
      ],
      "Resource": [
        "arn:aws:apigateway:${AWS_REGION}::/restapis",
        "arn:aws:apigateway:${AWS_REGION}::/restapis/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents",
        "logs:DescribeLogGroups",
        "logs:DeleteLogGroup"
      ],
      "Resource": "arn:aws:logs:${AWS_REGION}:${AWS_ACCOUNT_ID}:*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "iam:PassRole",
        "iam:GetRole",
        "iam:CreateRole",
        "iam:DeleteRole",
        "iam:PutRolePolicy",
        "iam:AttachRolePolicy",
        "iam:DetachRolePolicy"
      ],
      "Resource": [
        "arn:aws:iam::${AWS_ACCOUNT_ID}:role/cdk-*",
        "arn:aws:iam::${AWS_ACCOUNT_ID}:role/*-lambda-role"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "ssm:GetParameter"
      ],
      "Resource": "arn:aws:ssm:${AWS_REGION}:${AWS_ACCOUNT_ID}:parameter/cdk-bootstrap/*"
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

  return 0
}

# Attach policies to a role
attach_policies_to_role() {
  local role_name=$1
  
  echo -e "${BLUE}Attaching policies to ${role_name}...${NC}"
  
  # Create the deployment policy
  create_deployment_policy ${role_name}
  
  # Attach the policy to the role
  aws iam put-role-policy \
    --role-name ${role_name} \
    --policy-name CDKDeploymentPolicy \
    --policy-document file://${POLICY_DIR}/policy-${role_name}.json
  
  echo -e "${GREEN}Policies attached to ${role_name}.${NC}"
}

# Create or update IAM role
create_or_update_role() {
  local role_name=$1
  local policy_file=$2
  local description=$3
  
  echo -e "${BLUE}Creating or updating IAM role: ${role_name}...${NC}"
  
  local role_arn=""
  
  if aws iam get-role --role-name ${role_name} &> /dev/null; then
    echo -e "${BLUE}Role ${role_name} already exists. Updating...${NC}"
    
    # Update trust policy
    echo -e "${BLUE}Updating trust policy for ${role_name}...${NC}"
    aws iam update-assume-role-policy \
      --role-name ${role_name} \
      --policy-document file://${policy_file}
    
    role_arn=$(aws iam get-role --role-name ${role_name} --query "Role.Arn" --output text)
  else
    echo -e "${BLUE}Creating new ${role_name} role...${NC}"
    aws iam create-role \
      --role-name ${role_name} \
      --description "${description}" \
      --assume-role-policy-document file://${policy_file}
    
    role_arn=$(aws iam get-role --role-name ${role_name} --query "Role.Arn" --output text)
  fi
  
  # Attach policies
  attach_policies_to_role ${role_name}
  
  echo -e "${GREEN}Role ${role_name} updated successfully.${NC}"
  
  # Return the role ARN
  echo ${role_arn}
}

# Create or update IAM roles for GitHub Actions
create_iam_roles() {
  echo -e "${BLUE}Creating or updating IAM roles for GitHub Actions...${NC}"
  
  # Create or update role for development environment
  DEVELOPMENT_ROLE_ARN=$(create_or_update_role "GitHubActionsDevelopment" \
    "${POLICY_DIR}/trust-policy-development.json" \
    "Role for GitHub Actions development environment")
  
  # Create or update role for staging environment
  STAGING_ROLE_ARN=$(create_or_update_role "GitHubActionsStaging" \
    "${POLICY_DIR}/trust-policy-staging.json" \
    "Role for GitHub Actions staging environment")
  
  # Create or update role for production environment
  PRODUCTION_ROLE_ARN=$(create_or_update_role "GitHubActionsProduction" \
    "${POLICY_DIR}/trust-policy-production.json" \
    "Role for GitHub Actions production environment")
  
  echo -e "${GREEN}All IAM roles created or updated successfully.${NC}"
}

# Set up GitHub repository secrets
setup_github_secrets() {
  echo -e "${BLUE}Setting up GitHub repository secrets...${NC}"
  
  if [ "$HAS_GH" = true ]; then
    # Set secrets using GitHub CLI
    echo -e "${BLUE}Setting AWS_ROLE_TO_ASSUME_DEVELOPMENT secret...${NC}"
    gh secret set AWS_ROLE_TO_ASSUME_DEVELOPMENT -b"$DEVELOPMENT_ROLE_ARN"
    
    echo -e "${BLUE}Setting AWS_ROLE_TO_ASSUME_STAGING secret...${NC}"
    gh secret set AWS_ROLE_TO_ASSUME_STAGING -b"$STAGING_ROLE_ARN"
    
    echo -e "${BLUE}Setting AWS_ROLE_TO_ASSUME_PRODUCTION secret...${NC}"
    gh secret set AWS_ROLE_TO_ASSUME_PRODUCTION -b"$PRODUCTION_ROLE_ARN"
    
    # For backward compatibility, also set the generic role secrets
    echo -e "${BLUE}Setting AWS_ROLE_TO_ASSUME secret (for backward compatibility)...${NC}"
    gh secret set AWS_ROLE_TO_ASSUME -b"$DEVELOPMENT_ROLE_ARN"
    
    echo -e "${BLUE}Setting AWS_ROLE_TO_ASSUME_PROD secret (for backward compatibility)...${NC}"
    gh secret set AWS_ROLE_TO_ASSUME_PROD -b"$PRODUCTION_ROLE_ARN"
    
    echo -e "${BLUE}Setting AWS_REGION secret...${NC}"
    gh secret set AWS_REGION -b"$AWS_REGION"
    
    echo -e "${GREEN}GitHub secrets set successfully.${NC}"
  else
    # Provide instructions for manual setup
    echo -e "${YELLOW}GitHub CLI not available. Please set up the following secrets manually:${NC}"
    echo -e "${YELLOW}1. Go to https://github.com/${GITHUB_OWNER}/${GITHUB_REPO_NAME}/settings/secrets/actions${NC}"
    echo -e "${YELLOW}2. Add the following secrets:${NC}"
    echo -e "${YELLOW}   - AWS_ROLE_TO_ASSUME_DEVELOPMENT: ${DEVELOPMENT_ROLE_ARN}${NC}"
    echo -e "${YELLOW}   - AWS_ROLE_TO_ASSUME_STAGING: ${STAGING_ROLE_ARN}${NC}"
    echo -e "${YELLOW}   - AWS_ROLE_TO_ASSUME_PRODUCTION: ${PRODUCTION_ROLE_ARN}${NC}"
    echo -e "${YELLOW}   - AWS_ROLE_TO_ASSUME: ${DEVELOPMENT_ROLE_ARN} (for backward compatibility)${NC}"
    echo -e "${YELLOW}   - AWS_ROLE_TO_ASSUME_PROD: ${PRODUCTION_ROLE_ARN} (for backward compatibility)${NC}"
    echo -e "${YELLOW}   - AWS_REGION: ${AWS_REGION}${NC}"
  fi
}

# Clean up temporary files
cleanup() {
  echo -e "${BLUE}Cleaning up temporary files...${NC}"
  
  # Uncomment if you want to remove the policy files after setup
  # rm -rf ${POLICY_DIR}
  
  echo -e "${GREEN}Cleanup complete.${NC}"
}

# Main function
main() {
  echo -e "${BLUE}Setting up GitHub Actions with AWS IAM Identity Federation...${NC}"
  
  check_requirements
  get_aws_account_id
  get_aws_region
  get_github_repo_info
  create_oidc_provider
  create_trust_policies
  create_iam_roles
  setup_github_secrets
  # cleanup  # Uncomment if you want to clean up temporary files
  
  echo -e "${GREEN}Setup completed successfully!${NC}"
  echo -e "${GREEN}Your GitHub Actions workflows are now ready to use AWS IAM Identity Federation.${NC}"
  echo -e "${GREEN}Role ARNs:${NC}"
  echo -e "${GREEN}  Development: ${DEVELOPMENT_ROLE_ARN}${NC}"
  echo -e "${GREEN}  Staging: ${STAGING_ROLE_ARN}${NC}"
  echo -e "${GREEN}  Production: ${PRODUCTION_ROLE_ARN}${NC}"
  echo -e "${BLUE}The following GitHub secrets have been set:${NC}"
  echo -e "${BLUE}  - AWS_ROLE_TO_ASSUME_DEVELOPMENT${NC}"
  echo -e "${BLUE}  - AWS_ROLE_TO_ASSUME_STAGING${NC}"
  echo -e "${BLUE}  - AWS_ROLE_TO_ASSUME_PRODUCTION${NC}"
  echo -e "${BLUE}  - AWS_ROLE_TO_ASSUME (for backward compatibility)${NC}"
  echo -e "${BLUE}  - AWS_ROLE_TO_ASSUME_PROD (for backward compatibility)${NC}"
  echo -e "${BLUE}  - AWS_REGION${NC}"
}

# Run the main function
main
