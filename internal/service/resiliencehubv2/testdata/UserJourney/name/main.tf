# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_resiliencehubv2_user_journey" "test" {
  name       = var.name
  system_arn = aws_resiliencehubv2_system.test.arn
}

resource "aws_resiliencehubv2_system" "test" {
  name = "${var.rName}-system"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "name" {
  type     = string
  nullable = false
}
