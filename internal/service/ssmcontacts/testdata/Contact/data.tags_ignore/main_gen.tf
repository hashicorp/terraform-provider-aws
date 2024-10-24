# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "aws" {
  default_tags {
    tags = var.provider_tags
  }
  ignore_tags {
    keys = var.ignore_tag_keys
  }
}

# tflint-ignore: terraform_unused_declarations
data "aws_ssmcontacts_contact" "test" {
  arn = aws_ssmcontacts_contact.test.arn
}

resource "aws_ssmcontacts_contact" "test" {
  alias = var.rName
  type  = "PERSONAL"

  tags = var.resource_tags

  depends_on = [aws_ssmincidents_replication_set.test]
}

# testAccContactConfig_base

resource "aws_ssmincidents_replication_set" "test" {
  region {
    name = data.aws_region.current.name
  }
}

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

variable "provider_tags" {
  type     = map(string)
  nullable = true
  default  = null
}

variable "ignore_tag_keys" {
  type     = set(string)
  nullable = false
}
