# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_efs_backup_policy" "test" {
  region = var.region

  file_system_id = aws_efs_file_system.test.id

  backup_policy {
    status = "ENABLED"
  }
}

resource "aws_efs_file_system" "test" {
  region = var.region

}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
