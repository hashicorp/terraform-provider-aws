# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole",
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}",
      }
      Effect = "Allow"
      Sid    = ""
    }]
  })
}

data "aws_partition" "current" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
