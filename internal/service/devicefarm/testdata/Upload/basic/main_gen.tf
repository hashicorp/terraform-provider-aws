# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_devicefarm_upload" "test" {
  name        = var.rName
  project_arn = aws_devicefarm_project.test.arn
  type        = "APPIUM_JAVA_TESTNG_TEST_SPEC"
}

resource "aws_devicefarm_project" "test" {
  name = var.rName
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
