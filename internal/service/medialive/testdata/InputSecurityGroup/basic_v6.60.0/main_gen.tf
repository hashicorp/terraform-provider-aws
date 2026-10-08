# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_medialive_input_security_group" "test" {
  whitelist_rules {
    cidr = "10.2.0.0/16"
  }
}

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.60.0"
    }
  }
}

provider "aws" {}
