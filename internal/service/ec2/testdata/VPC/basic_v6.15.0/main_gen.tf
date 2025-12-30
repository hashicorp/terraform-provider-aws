# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
}

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.15.0"
    }
  }
}

provider "aws" {}
