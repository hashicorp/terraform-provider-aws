# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_resiliencehubv2_policy" "test" {
  name = var.rName

  multi_region {
    disaster_recovery_approach = "HOT_STANDBY"
    rpo_in_minutes             = var.rpo_in_minutes
    rto_in_minutes             = var.rto_in_minutes
  }
}

variable "rpo_in_minutes" {
  type     = number
  nullable = true
  default  = null
}

variable "rto_in_minutes" {
  type     = number
  nullable = true
  default  = null
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
