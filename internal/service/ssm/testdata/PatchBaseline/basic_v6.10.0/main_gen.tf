# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

resource "aws_ssm_patch_baseline" "test" {
  name                              = var.rName
  description                       = "Baseline containing all updates approved for production systems"
  approved_patches                  = ["KB123456"]
  approved_patches_compliance_level = "CRITICAL"
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
