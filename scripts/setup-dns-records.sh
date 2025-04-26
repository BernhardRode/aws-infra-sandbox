#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Domain to configure
DOMAIN="ebbo.dev"

# Check if required tools are installed
check_requirements() {
  echo -e "${BLUE}Checking requirements for DNS setup...${NC}"
  
  # Check AWS CLI
  if ! command -v aws &> /dev/null; then
    echo -e "${RED}AWS CLI is not installed. Please install it first.${NC}"
    echo "Visit: https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html"
    exit 1
  fi
  
  # Check if AWS CLI is configured
  if ! aws sts get-caller-identity &> /dev/null; then
    echo -e "${RED}AWS CLI is not configured properly. Please run 'aws configure'.${NC}"
    exit 1
  fi
  
  echo -e "${GREEN}All required tools are available for DNS setup.${NC}"
}

# Check if domain exists in Route 53
check_domain() {
  echo -e "${BLUE}Checking if domain ${DOMAIN} exists in Route 53...${NC}"
  
  # Try to get the hosted zone ID for the domain
  HOSTED_ZONE_ID=$(aws route53 list-hosted-zones-by-name --dns-name "${DOMAIN}." --max-items 1 --query "HostedZones[?Name=='${DOMAIN}.'].Id" --output text)
  
  if [ -z "$HOSTED_ZONE_ID" ]; then
    echo -e "${RED}Domain ${DOMAIN} not found in Route 53.${NC}"
    echo -e "${YELLOW}Please create the hosted zone for ${DOMAIN} before running this script.${NC}"
    echo -e "${YELLOW}You can create it with: aws route53 create-hosted-zone --name ${DOMAIN} --caller-reference $(date +%s)${NC}"
    exit 1
  else
    # Extract just the zone ID without the /hostedzone/ prefix
    HOSTED_ZONE_ID=$(echo $HOSTED_ZONE_ID | sed 's/\/hostedzone\///')
    echo -e "${GREEN}Domain ${DOMAIN} found in Route 53 with Hosted Zone ID: ${HOSTED_ZONE_ID}${NC}"
  fi
}

# Add A and AAAA records
add_dns_records() {
  echo -e "${BLUE}Adding DNS records for ${DOMAIN}...${NC}"
  
  # A Records - IPv4 addresses
  IPV4_ADDRESSES=(
    "76.76.21.21"
    "76.76.21.22"
    "76.76.21.23"
    "76.76.21.24"
  )
  
  # AAAA Records - IPv6 addresses
  IPV6_ADDRESSES=(
    "2606:1a40:1:7::1"
    "2606:1a40:1:7::2"
    "2606:1a40:1:7::3"
    "2606:1a40:1:7::4"
  )
  
  # Create change batch file for Route 53
  TEMP_FILE=$(mktemp)
  
  echo '{
    "Comment": "Adding A and AAAA records for '"${DOMAIN}"'",
    "Changes": [' > $TEMP_FILE
  
  # Add A records
  for i in "${!IPV4_ADDRESSES[@]}"; do
    echo '      {
        "Action": "UPSERT",
        "ResourceRecordSet": {
          "Name": "'"${DOMAIN}"'",
          "Type": "A",
          "TTL": 300,
          "ResourceRecords": [
            {
              "Value": "'"${IPV4_ADDRESSES[$i]}"'"
            }
          ]
        }
      }' >> $TEMP_FILE
    
    # Add comma if not the last item and if there are AAAA records
    if [ $i -lt $((${#IPV4_ADDRESSES[@]} - 1)) ] || [ ${#IPV6_ADDRESSES[@]} -gt 0 ]; then
      echo ',' >> $TEMP_FILE
    fi
  done
  
  # Add AAAA records
  for i in "${!IPV6_ADDRESSES[@]}"; do
    echo '      {
        "Action": "UPSERT",
        "ResourceRecordSet": {
          "Name": "'"${DOMAIN}"'",
          "Type": "AAAA",
          "TTL": 300,
          "ResourceRecords": [
            {
              "Value": "'"${IPV6_ADDRESSES[$i]}"'"
            }
          ]
        }
      }' >> $TEMP_FILE
    
    # Add comma if not the last item
    if [ $i -lt $((${#IPV6_ADDRESSES[@]} - 1)) ]; then
      echo ',' >> $TEMP_FILE
    fi
  done
  
  echo '
    ]
  }' >> $TEMP_FILE
  
  # Apply the changes
  echo -e "${BLUE}Applying DNS changes to Route 53...${NC}"
  aws route53 change-resource-record-sets --hosted-zone-id $HOSTED_ZONE_ID --change-batch file://$TEMP_FILE
  
  # Clean up
  rm $TEMP_FILE
  
  echo -e "${GREEN}DNS records added successfully for ${DOMAIN}.${NC}"
}

# Main function
main() {
  echo -e "${BLUE}Setting up DNS records for ${DOMAIN}...${NC}"
  
  check_requirements
  check_domain
  add_dns_records
  
  echo -e "${GREEN}DNS setup completed successfully!${NC}"
}

# Run the main function
main
