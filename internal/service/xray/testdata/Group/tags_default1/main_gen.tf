# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "aws" {
  default_tags {
    tags = {
      (var.providerTagKey1) = var.providerTagValue1
    }
  }
}

resource "aws_xray_group" "test" {
  group_name        = var.rName
  filter_expression = "responsetime > 5"

  tags = var.tags
}


variable "rName" {
  type     = string
  nullable = false
}

variable "tags" {
  type     = map(string)
  nullable = false
}


variable "providerTagKey1" {
  type     = string
  nullable = false
}

variable "providerTagValue1" {
  type     = string
  nullable = false
}
