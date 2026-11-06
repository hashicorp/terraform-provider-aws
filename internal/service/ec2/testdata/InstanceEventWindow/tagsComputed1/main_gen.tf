# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_ec2_instance_event_window" "test" {
  name = var.rName

  time_ranges {
    start_week_day = "sunday"
    start_hour     = 2
    end_week_day   = "sunday"
    end_hour       = 6
  }

  tags = {
    (var.unknownTagKey) = null_resource.test.id
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
