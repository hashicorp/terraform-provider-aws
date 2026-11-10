# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_resiliencehubv2_policy" "test" {
  name = var.rName

  data_recovery {
    time_between_backups_in_minutes = var.time_between_backups_in_minutes
  }
}

variable "time_between_backups_in_minutes" {
  type     = number
  nullable = false
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
