# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "aws" {
  default_tags {
    tags = var.provider_tags
  }
}

data "aws_caller_identity" "current" {}

resource "aws_fms_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id
}

resource "aws_fms_resource_set" "test" {
  depends_on = [aws_fms_admin_account.test]
  resource_set {
    name               = var.rName
    resource_type_list = ["AWS::NetworkFirewall::Firewall"]
  }

  tags = var.resource_tags
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
