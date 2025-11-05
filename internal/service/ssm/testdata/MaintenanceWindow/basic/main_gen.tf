# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_ssm_maintenance_window" "test" {
  name     = var.rName
  cutoff   = 1
  duration = 3
  schedule = "cron(0 16 ? * TUE *)"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
