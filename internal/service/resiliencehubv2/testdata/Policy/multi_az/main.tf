# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_resiliencehubv2_policy" "test" {
  name = var.rName

  multi_az {
    disaster_recovery_approach = "ACTIVE_ACTIVE"
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
