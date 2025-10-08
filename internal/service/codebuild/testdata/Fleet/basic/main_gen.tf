# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_codebuild_fleet" "test" {
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
