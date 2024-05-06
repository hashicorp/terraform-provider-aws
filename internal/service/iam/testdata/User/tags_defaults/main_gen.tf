# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "aws" {
  default_tags {
    tags = var.provider_tags
  }
}

resource "aws_iam_user" "test" {
  name = var.rName

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


variable "provider_tags" {
  type     = map(string)
  nullable = false
}
