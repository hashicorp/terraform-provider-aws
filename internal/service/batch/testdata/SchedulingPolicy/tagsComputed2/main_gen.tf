# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_batch_scheduling_policy" "test" {
  name = var.rName

  fair_share_policy {
    compute_reservation = 0
    share_decay_seconds = 0
  }

  tags = {
    (var.unknownTagKey) = null_resource.test.id
    (var.knownTagKey)   = var.knownTagValue
  }
}

resource "null_resource" "test" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "unknownTagKey" {
  type     = string
  nullable = false
}

variable "knownTagKey" {
  type     = string
  nullable = false
}

variable "knownTagValue" {
  type     = string
  nullable = false
}
