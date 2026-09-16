# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_efs_replication_configuration" "test" {
  region = var.region

  source_file_system_id = aws_efs_file_system.test.id

  destination {
    region = var.destination_region
  }
}

resource "aws_efs_file_system" "test" {
  region = var.region
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}

variable "destination_region" {
  type        = string
  nullable    = false
}