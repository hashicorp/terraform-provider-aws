# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_iam_user" "expected" {
  count = var.resource_count

  name = "${var.rName}-${count.index}"
  path = var.expected_path_name
}

resource "aws_iam_user" "not_expected" {
  count = var.resource_count

  name = "${var.rName}-other-${count.index}"
  path = var.other_path_name
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

variable "expected_path_name" {
  description = "Path name for expected resources"
  type        = string
  nullable    = false
}

variable "other_path_name" {
  description = "Path name for non-expected resources"
  type        = string
  nullable    = false
}
