# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_iam_role" "test" {
  count = var.resource_count

  name = "${var.rName}-${count.index}"

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

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}
