#!/bin/bash

# Define the provider and registry details
PROVIDER="hashicorp/aws"
REGISTRY_URL="https://registry.terraform.io/v2/providers/323/provider-versions/latest"


# Fetch the latest version of the AWS Terraform provider
LATEST_VERSION=$(curl -s $REGISTRY_URL | jq '.data.attributes.version')

# Compare this latest version to the github version

# Check if the latest version was retrieved
if [[ -z "$LATEST_VERSION" ]]; then
    echo "Failed to fetch the latest version of the AWS provider."
    exit 1
fi

echo "Latest AWS Terraform Provider version: $LATEST_VERSION"

# Optionally, you can set this version in a Terraform configuration file
echo "Setting provider version in Terraform configuration..."
cat <<EOF > main.tf
terraform {
  required_providers {
    aws = {
      source  = "${PROVIDER}"
      version = ${LATEST_VERSION}
    }
  }
}

provider "aws" {
  region = "us-east-1"  # Change this to your preferred region
}
EOF

echo "Terraform configuration file (main.tf) created with the latest version."

# Initialize Terraform and pull the latest provider
echo "Initializing Terraform and pulling the latest AWS provider..."
terraform init
