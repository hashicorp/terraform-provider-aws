# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_cloudwatch_alarm_mute_rule" "test" {
  name = var.rName

  rule {
    schedule {
      duration   = "PT4H"
      expression = "cron(0 2 * * *)"
    }
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
