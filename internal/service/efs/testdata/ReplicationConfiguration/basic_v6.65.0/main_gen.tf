# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_efs_replication_configuration" "test" {
  source_file_system_id = aws_efs_file_system.test.id

  destination {
    region = var.destination_region
  }
}

resource "aws_efs_file_system" "test" {}

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.65.0"
    }
  }
}

provider "aws" {}

variable "destination_region" {
  type        = string
  nullable    = false
}