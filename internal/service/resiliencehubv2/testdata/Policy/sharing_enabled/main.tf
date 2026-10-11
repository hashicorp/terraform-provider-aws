# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_resiliencehubv2_policy" "test" {
  name = var.rName

  availability_slo {
    target = 99.9
  }

  sharing_enabled = var.sharing_enabled
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "sharing_enabled" {
  type     = bool
  nullable = false
}
