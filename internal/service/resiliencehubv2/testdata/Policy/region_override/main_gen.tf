# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_resiliencehubv2_policy" "test" {
  region = var.region

  name = var.rName

  availability_slo {
    target = 99.9
  }

  multi_az {
    disaster_recovery_approach = "ACTIVE_ACTIVE"
    rpo_in_minutes             = 5
    rto_in_minutes             = 10
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
