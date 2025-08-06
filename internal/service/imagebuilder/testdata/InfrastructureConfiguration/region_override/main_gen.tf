# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  region = var.region

  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = var.rName
}

resource "aws_iam_instance_profile" "test" {
  name = var.rName
  role = aws_iam_role.test.name
}

resource "aws_iam_role" "test" {
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.amazonaws.com"
      }
    }]
  })
  name = var.rName
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
