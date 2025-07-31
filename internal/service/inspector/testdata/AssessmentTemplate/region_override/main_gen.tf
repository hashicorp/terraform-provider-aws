# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

data "aws_inspector_rules_packages" "available" {
  region = var.region

}

resource "aws_inspector_resource_group" "test" {
  region = var.region

  tags = {
    Name = var.rName
  }
}

resource "aws_inspector_assessment_target" "test" {
  region = var.region

  name               = var.rName
  resource_group_arn = aws_inspector_resource_group.test.arn
}

resource "aws_inspector_assessment_template" "test" {
  region = var.region

  name       = var.rName
  target_arn = aws_inspector_assessment_target.test.arn
  duration   = 3600

  rules_package_arns = data.aws_inspector_rules_packages.available.arns
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
