# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ec2_instance_event_window" "test" {
  region = var.region

  name = var.rName

  time_ranges {
    start_week_day = "sunday"
    start_hour     = 2
    end_week_day   = "sunday"
    end_hour       = 6
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
