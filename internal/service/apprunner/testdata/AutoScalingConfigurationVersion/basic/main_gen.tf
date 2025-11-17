# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_apprunner_auto_scaling_configuration_version" "test" {
  auto_scaling_configuration_name = var.rName
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
