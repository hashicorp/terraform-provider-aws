# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "aws" {
  default_tags {
    tags = var.provider_tags
  }
}

resource "aws_batch_scheduling_policy" "test" {
  name = var.rName

  fair_share_policy {
    compute_reservation = 0
    share_decay_seconds = 0
  }

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
