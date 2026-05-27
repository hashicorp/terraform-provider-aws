# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_rds_certificate" "test" {
  certificate_identifier = "rds-ca-rsa4096-g1"
}

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.0.0"
    }
  }
}

provider "aws" {}
