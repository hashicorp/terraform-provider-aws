# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "aws" {
  default_tags {
    tags = var.provider_tags
  }
}

resource "aws_xray_group" "test" {
  group_name        = var.rName
  filter_expression = "responsetime > 5"

}


variable "rName" {
  type     = string
  nullable = false
}


variable "provider_tags" {
  type     = map(string)
  nullable = false
}
