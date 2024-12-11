# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_appconfig_environment" "test" {
  name           = var.rName
  application_id = aws_appconfig_application.test.id

  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }
}

resource "aws_appconfig_application" "test" {
  name = var.rName
}

resource "null_resource" "test" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "unknownTagKey" {
  type     = string
  nullable = false
}
