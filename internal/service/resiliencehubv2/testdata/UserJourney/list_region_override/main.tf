# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_resiliencehubv2_system" "test" {
  region = var.region

  name = "${var.rName}-system"
}

resource "aws_resiliencehubv2_user_journey" "test" {
  count  = var.resource_count
  region = var.region

  name       = "${var.rName}-${count.index}"
  system_arn = aws_resiliencehubv2_system.test.arn
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
