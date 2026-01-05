# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

resource "aws_imagebuilder_distribution_configuration" "test" {
  region = var.region

  name = var.rName

  distribution {
    ami_distribution_configuration {
      name = "test-name-{{ imagebuilder:buildDate }}"
    }

    region = data.aws_region.current.name
  }
}

data "aws_region" "current" {
  region = var.region

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
