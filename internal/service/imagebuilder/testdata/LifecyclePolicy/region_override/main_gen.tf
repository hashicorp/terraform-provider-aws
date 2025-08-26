# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_imagebuilder_lifecycle_policy" "test" {
  region = var.region

  name           = var.rName
  description    = "Used for setting lifecycle policies"
  execution_role = aws_iam_role.test.arn
  resource_type  = "AMI_IMAGE"
  policy_detail {
    action {
      type = "DELETE"
    }
    filter {
      type            = "AGE"
      value           = 6
      retain_at_least = 10
      unit            = "YEARS"
    }
  }
  resource_selection {
    tag_map = {
      "key1" = "value1"
      "key2" = "value2"
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}

# testAccLifecyclePolicyConfig_base

resource "aws_iam_role" "test" {
  name = var.rName
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "imagebuilder.${data.aws_partition.current.dns_suffix}"
      }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/EC2ImageBuilderLifecycleExecutionPolicy"
  role       = aws_iam_role.test.name
}

data "aws_partition" "current" {}

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
