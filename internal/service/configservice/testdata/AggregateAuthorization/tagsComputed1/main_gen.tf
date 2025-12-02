# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_config_aggregate_authorization" "test" {
  account_id            = data.aws_caller_identity.current.account_id
  authorized_aws_region = data.aws_region.default.name

  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }
}

data "aws_caller_identity" "current" {}

data "aws_region" "default" {}

resource "null_resource" "test" {}

variable "unknownTagKey" {
  type     = string
  nullable = false
}
