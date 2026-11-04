# Copyright IBM Corp. 2014, 2026
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
data "aws_pinpointsmsvoicev2_sender_id" "test" {
  sender_id        = aws_pinpointsmsvoicev2_sender_id.test.sender_id
  iso_country_code = aws_pinpointsmsvoicev2_sender_id.test.iso_country_code
}

resource "aws_pinpointsmsvoicev2_sender_id" "test" {
  sender_id        = var.rName
  iso_country_code = "GB"

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
  nullable = true
  default  = null
}

variable "ignore_tag_keys" {
  type     = set(string)
  nullable = false
}
