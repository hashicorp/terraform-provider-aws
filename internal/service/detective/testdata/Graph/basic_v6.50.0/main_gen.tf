# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_detective_graph" "test" {
}
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.50.0"
    }
  }
}

provider "aws" {}
