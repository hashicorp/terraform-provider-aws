# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

data "aws_resiliencehubv2_policy" "test" {
  arn = aws_resiliencehubv2_policy.test.arn
}

resource "aws_resiliencehubv2_policy" "test" {
  name = var.rName

  availability_slo {
    target = 99.9
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
