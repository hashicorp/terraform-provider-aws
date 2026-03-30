# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_devicefarm_device_pool" "test" {
  name        = var.rName
  project_arn = aws_devicefarm_project.test.arn
  rule {
    attribute = "OS_VERSION"
    operator  = "EQUALS"
    value     = "\"AVAILABLE\""
  }
}

# testAccProjectConfig_basic

resource "aws_devicefarm_project" "test" {
  name = var.rName
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
