# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatch_log_group" "test" {
  name = var.rName

  retention_in_days = 1
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.7.0"
    }
  }
}

provider "aws" {}
