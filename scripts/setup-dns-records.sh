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
TTL=300

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
    "185.199.108.153"
    "185.199.109.153"
    "185.199.110.153"
    "185.199.111.153"
  )
  
  # AAAA Records - IPv6 addresses
  IPV6_ADDRESSES=(
    "2606:50c0:8000::153"
    "2606:50c0:8001::153"
    "2606:50c0:8002::153"
    "2606:50c0:8003::153"
  )
  
  # Create change batch file for Route 53
  TEMP_FILE=$(mktemp)
  
  echo '{
    "Comment": "Adding A and AAAA records for '"${DOMAIN}"'",
    "Changes": [' > $TEMP_FILE
  
  # Add A records (all in one resource record set)
  echo '      {
        "Action": "UPSERT",
        "ResourceRecordSet": {
          "Name": "'"${DOMAIN}"'",
          "Type": "A",
          "TTL": '"$TTL"',
          "ResourceRecords": [' >> $TEMP_FILE
  
  # Add all IPv4 addresses
  for i in "${!IPV4_ADDRESSES[@]}"; do
    echo '            {
              "Value": "'"${IPV4_ADDRESSES[$i]}"'"
            }' >> $TEMP_FILE
    
    # Add comma if not the last item
    if [ $i -lt $((${#IPV4_ADDRESSES[@]} - 1)) ]; then
      echo ',' >> $TEMP_FILE
    fi
  done
  
  echo '          ]
        }
      }' >> $TEMP_FILE
  
  # Add comma between A and AAAA records if AAAA records exist
  if [ ${#IPV6_ADDRESSES[@]} -gt 0 ]; then
    echo ',' >> $TEMP_FILE
  fi
  
  # Add AAAA records (all in one resource record set)
  if [ ${#IPV6_ADDRESSES[@]} -gt 0 ]; then
    echo '      {
        "Action": "UPSERT",
        "ResourceRecordSet": {
          "Name": "'"${DOMAIN}"'",
          "Type": "AAAA",
          "TTL": '"$TTL"',
          "ResourceRecords": [' >> $TEMP_FILE
    
    # Add all IPv6 addresses
    for i in "${!IPV6_ADDRESSES[@]}"; do
      echo '            {
              "Value": "'"${IPV6_ADDRESSES[$i]}"'"
            }' >> $TEMP_FILE
      
      # Add comma if not the last item
      if [ $i -lt $((${#IPV6_ADDRESSES[@]} - 1)) ]; then
        echo ',' >> $TEMP_FILE
      fi
    done
    
    echo '          ]
        }
      }' >> $TEMP_FILE
  fi
  
  echo '
    ]
  }' >> $TEMP_FILE
  
  # Debug: Show the change batch file
  echo -e "${BLUE}Change batch file content:${NC}"
  cat $TEMP_FILE
  
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
