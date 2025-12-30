# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

resource "aws_ssm_maintenance_window" "test" {
  region = var.region

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

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
