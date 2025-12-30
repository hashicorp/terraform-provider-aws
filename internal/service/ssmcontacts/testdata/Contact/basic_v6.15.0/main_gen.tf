# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

resource "aws_ssmcontacts_contact" "test" {
  alias = var.rName
  type  = "PERSONAL"

  depends_on = [aws_ssmincidents_replication_set.test]
}

# testAccContactConfig_base

resource "aws_ssmincidents_replication_set" "test" {
  region {
    name = data.aws_region.current.region
  }
}

data "aws_region" "current" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
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
