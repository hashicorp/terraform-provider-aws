# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_flow_log" "test" {
  region = var.region

  iam_role_arn         = aws_iam_role.test.arn
  log_destination      = aws_cloudwatch_log_group.test.arn
  log_destination_type = "cloud-watch-logs"
  traffic_type         = "ALL"
  vpc_id               = aws_vpc.test.id
}

data "aws_partition" "test" {}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        Service = "ec2.${data.aws_partition.test.dns_suffix}"
      }
      Action = "sts:AssumeRole"
    }]
  })
}

resource "aws_cloudwatch_log_group" "test" {
  region = var.region

  name = var.rName
}

resource "aws_vpc" "test" {
  region = var.region

  cidr_block = "10.1.0.0/16"
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
