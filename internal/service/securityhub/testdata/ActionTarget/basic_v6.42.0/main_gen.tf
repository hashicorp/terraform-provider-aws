# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_securityhub_account" "test" {
}

resource "aws_securityhub_action_target" "test" {
  depends_on  = [aws_securityhub_account.test]
  description = "description1"
  identifier  = "testaction"
  name        = "Test action"
}

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.42.0"
    }
  }
}

provider "aws" {}
