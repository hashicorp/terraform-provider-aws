# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_codebuild_fleet" "test" {
  region = var.region

  base_capacity     = 1
  compute_type      = "BUILD_GENERAL1_SMALL"
  environment_type  = "LINUX_CONTAINER"
  name              = var.rName
  overflow_behavior = "ON_DEMAND"
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
