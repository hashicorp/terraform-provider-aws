# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_config_configuration_aggregator" "test" {
  name = var.rName

  account_aggregation_source {
    account_ids = [data.aws_caller_identity.current.account_id]
    regions     = [data.aws_region.current.region]
  }

  tags = var.resource_tags
}

data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

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
