# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_config_aggregate_authorization" "test" {
  account_id            = data.aws_caller_identity.current.account_id
  authorized_aws_region = data.aws_region.default.name

  tags = var.resource_tags
}

data "aws_caller_identity" "current" {}

data "aws_region" "default" {}

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}
