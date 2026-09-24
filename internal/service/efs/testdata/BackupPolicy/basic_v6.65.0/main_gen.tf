# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_efs_backup_policy" "test" {
  file_system_id = aws_efs_file_system.test.id

  backup_policy {
    status = "ENABLED"
  }
}

resource "aws_efs_file_system" "test" {
}

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.65.0"
    }
  }
}

provider "aws" {}
