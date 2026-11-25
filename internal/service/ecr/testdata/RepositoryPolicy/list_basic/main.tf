# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ecr_repository_policy" "test" {
  count = var.resource_count

  repository = aws_ecr_repository.test[count.index].name

  policy = jsonencode({
    Version = "2008-10-17"
    Statement = [{
      Sid       = "${var.rName}-${count.index}"
      Effect    = "Allow"
      Principal = "*"
      Action    = "ecr:ListImages"
    }]
  })
}

resource "aws_ecr_repository" "test" {
  count = var.resource_count

  name = "${var.rName}-${count.index}"
}

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
