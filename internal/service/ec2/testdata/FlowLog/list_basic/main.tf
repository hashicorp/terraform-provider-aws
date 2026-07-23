# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_flow_log" "test" {
  count = var.resource_count

  vpc_id               = aws_vpc.test[count.index].id
  traffic_type         = "ALL"
  iam_role_arn         = aws_iam_role.test[count.index].arn
  log_destination      = aws_cloudwatch_log_group.test[count.index].arn
  log_destination_type = "cloud-watch-logs"
}

resource "aws_vpc" "test" {
  count = var.resource_count

  cidr_block = "10.${count.index}.0.0/16"

  tags = {
    Name = "${var.rName}-${count.index}"
  }
}

resource "aws_cloudwatch_log_group" "test" {
  count = var.resource_count

  name = "${var.rName}-${count.index}"
}

resource "aws_iam_role" "test" {
  count = var.resource_count

  name = "${var.rName}-${count.index}"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid    = ""
      Effect = "Allow"
      Principal = {
        Service = "vpc-flow-logs.amazonaws.com"
      }
      Action = "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_role_policy" "test" {
  count = var.resource_count

  name = "${var.rName}-${count.index}"
  role = aws_iam_role.test[count.index].id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents",
        "logs:DescribeLogGroups",
        "logs:DescribeLogStreams",
      ]
      Resource = "*"
    }]
  })
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

