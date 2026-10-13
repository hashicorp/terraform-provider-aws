# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_sfn_state_machine" "test" {
  count = var.resource_count

  name     = "${var.rName}-${count.index}"
  role_arn = aws_iam_role.test.arn

  definition = <<EOF
{
  "Comment": "A Hello World example using a Pass state",
  "StartAt": "HelloWorld",
  "States": {
    "HelloWorld": {
      "Type": "Pass",
      "End": true
    }
  }
}
EOF

  tags = var.resource_tags
}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Service": "states.${data.aws_region.current.region}.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }]
}
EOF
}

data "aws_region" "current" {}

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

variable "resource_tags" {
  description = "Tags to set on resource"
  type        = map(string)
  nullable    = false
}
