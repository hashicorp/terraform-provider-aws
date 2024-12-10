# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "awsalternate" {
  region = var.alt_region
}

resource "aws_kms_replica_external_key" "test" {
  description             = var.rName
  enabled                 = true
  primary_key_arn         = aws_kms_external_key.test.arn
  deletion_window_in_days = 7

  key_material_base64 = "Wblj06fduthWggmsT0cLVoIMOkeLbc2kVfMud77i/JY="

  tags = var.resource_tags
}

# ACCEPTANCE TESTING ONLY -- NEVER EXPOSE YOUR KEY MATERIAL
resource "aws_kms_external_key" "test" {
  provider = awsalternate

  description  = "${var.rName}-source"
  multi_region = true
  enabled      = true

  key_material_base64 = "Wblj06fduthWggmsT0cLVoIMOkeLbc2kVfMud77i/JY="

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

variable "alt_region" {
  description = "Region for provider awsalternate"
  type        = string
  nullable    = false
}
