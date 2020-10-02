terraform {
  required_providers {
    aws = {
      version = "99.99.99"
    }
  }
}

provider "aws" {
  region = "us-west-2"
}

resource "aws_organizations_organization" "test" {
  enabled_policy_types = ["SERVICE_CONTROL_POLICY"]
}

resource "aws_organizations_policy" "test" {
  name    = "FullAWSAccess"
  content = ""
  
  depends_on = [aws_organizations_organization.test]
}
