# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_dx_gateway" "test" {
  name            = var.rName
  amazon_side_asn = var.rBgpAsn
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
variable "rBgpAsn" {
  type     = string
  nullable = false
}

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "5.100.0"
    }
  }
}

provider "aws" {}
