# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

resource "aws_ssm_maintenance_window" "test" {
  name     = var.rName
  cutoff   = 1
  duration = 3
  schedule = "cron(0 16 ? * TUE *)"
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
      version = "6.10.0"
    }
  }
}

provider "aws" {}
