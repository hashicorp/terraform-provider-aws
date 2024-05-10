# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "aws" {
  default_tags {
    tags = var.provider_tags
  }
}

resource "aws_iam_service_linked_role" "test" {
  aws_service_name = "autoscaling.amazonaws.com"
  custom_suffix    = var.rName

  tags = {
    (var.tagKey1) = null
  }
}


variable "rName" {
  type     = string
  nullable = false
}

variable "tagKey1" {
  type     = string
  nullable = false
}

variable "provider_tags" {
  type     = map(string)
  nullable = false
}
