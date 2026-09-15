# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_securityhub_finding_aggregator" "test" {
  linking_mode = "ALL_REGIONS"

  depends_on = [aws_securityhub_account.test]
}

resource "aws_securityhub_account" "test" {
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
