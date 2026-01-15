# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

data "aws_inspector_rules_packages" "available" {
}

resource "aws_inspector_resource_group" "test" {
  tags = {
    Name = var.rName
  }
}

resource "aws_inspector_assessment_target" "test" {
  name               = var.rName
  resource_group_arn = aws_inspector_resource_group.test.arn
}

resource "aws_inspector_assessment_template" "test" {
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
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.4.0"
    }
  }
}

provider "aws" {}
