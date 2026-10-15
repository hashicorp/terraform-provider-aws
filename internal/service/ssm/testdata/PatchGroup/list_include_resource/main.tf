# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ssm_patch_group" "test" {
  count       = var.resource_count
  baseline_id = aws_ssm_patch_baseline.test.id
  patch_group = "${var.rName}-${count.index}"
}

resource "aws_ssm_patch_baseline" "test" {
  name             = var.rName
  approved_patches = ["KB123456"]
}

variable "rName" {
  type = string
}

variable "resource_count" {
  type = number
}
