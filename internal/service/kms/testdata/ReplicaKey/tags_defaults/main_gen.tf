# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "aws" {
  default_tags {
    tags = var.provider_tags
  }
}

provider "awsalternate" {
  region = var.alt_region
}

resource "aws_kms_replica_key" "test" {
  description             = var.rName
  primary_key_arn         = aws_kms_key.test.arn
  deletion_window_in_days = 7

  tags = var.resource_tags
}

resource "aws_kms_key" "test" {
  provider = awsalternate

  description  = "${var.rName}-source"
  multi_region = true

  deletion_window_in_days = 7
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}

variable "provider_tags" {
  type     = map(string)
  nullable = false
}

variable "alt_region" {
  description = "Region for provider awsalternate"
  type        = string
  nullable    = false
}
