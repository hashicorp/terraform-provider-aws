# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = var.rName

  distribution {
    ami_distribution_configuration {
      name = "test-name-{{ imagebuilder:buildDate }}"
    }

    region = data.aws_region.current.name
  }
}

data "aws_region" "current" {
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
