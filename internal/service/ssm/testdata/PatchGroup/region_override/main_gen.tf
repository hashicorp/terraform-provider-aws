# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ssm_patch_baseline" "test" {
  region = var.region

  name             = var.rName
  approved_patches = ["KB123456"]
}

resource "aws_ssm_patch_group" "test" {
  region = var.region

  baseline_id = aws_ssm_patch_baseline.test.id
  patch_group = var.rName
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
