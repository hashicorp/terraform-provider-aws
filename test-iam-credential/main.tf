terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "999.999.999" # Use the local development version
    }
  }
}

provider "aws" {
  region = "us-east-1"
}

# Create an IAM user
resource "aws_iam_user" "test" {
  name = "test-service-credential-user"
}

# Create a CodeCommit credential
resource "aws_iam_service_specific_credential" "codecommit" {
  service_name = "codecommit.amazonaws.com"
  user_name    = aws_iam_user.test.name
}

# Create a Bedrock API key with expiration
resource "aws_iam_service_specific_credential" "bedrock" {
  service_name        = "bedrock.amazonaws.com"
  user_name           = aws_iam_user.test.name
  credential_age_days = 30
}

# Outputs
output "codecommit_password" {
  value     = aws_iam_service_specific_credential.codecommit.service_password
  sensitive = true
}

output "codecommit_username" {
  value = aws_iam_service_specific_credential.codecommit.service_user_name
}

output "bedrock_alias" {
  value = aws_iam_service_specific_credential.bedrock.service_credential_alias
}

output "bedrock_secret" {
  value     = aws_iam_service_specific_credential.bedrock.service_credential_secret
  sensitive = true
}

output "bedrock_expiration" {
  value = aws_iam_service_specific_credential.bedrock.expiration_date
}