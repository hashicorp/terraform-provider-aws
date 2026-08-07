# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_resiliencehubv2_policy" "test" {
  name = var.rName

  availability_slo {
    target = 99.95
  }

  data_recovery {
    time_between_backups_in_minutes = 20
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
