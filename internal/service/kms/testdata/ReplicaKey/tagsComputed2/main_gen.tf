# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_kms_replica_key" "test" {
  description             = var.rName
  primary_key_arn         = aws_kms_key.test.arn
  deletion_window_in_days = 7

  tags = {
    (var.unknownTagKey) = null_resource.test.id
    (var.knownTagKey)   = var.knownTagValue
  }
}

resource "aws_kms_key" "test" {
  region = var.secondary_region

  description  = "${var.rName}-source"
  multi_region = true

  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "null_resource" "test" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "unknownTagKey" {
  type     = string
  nullable = false
}

variable "knownTagKey" {
  type     = string
  nullable = false
}

variable "knownTagValue" {
  type     = string
  nullable = false
}

variable "secondary_region" {
  description = "Secondary region"
  type        = string
  nullable    = false
}
