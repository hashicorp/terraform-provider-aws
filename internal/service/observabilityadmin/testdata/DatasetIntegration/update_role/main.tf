# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_observabilityadmin_dataset_integration" "test" {
  role_arn = aws_iam_role.test[var.role_index].arn
}

resource "aws_iam_role" "test" {
  count = 2

  name = "${var.rName}-${count.index}"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "logs.amazonaws.com"
      }
    }]
  })
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "role_index" {
  description = "Index of the IAM role to associate with the dataset integration"
  type        = number
  nullable    = false
}
