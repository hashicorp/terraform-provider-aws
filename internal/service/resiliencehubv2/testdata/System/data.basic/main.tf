# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

# tflint-ignore: terraform_unused_declarations
data "aws_resiliencehubv2_system" "test" {
  arn = aws_resiliencehubv2_system.test.arn
}

resource "aws_resiliencehubv2_system" "test" {
  name = var.rName
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
