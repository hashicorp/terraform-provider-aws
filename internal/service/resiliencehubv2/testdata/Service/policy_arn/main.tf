# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

data "aws_region" "current" {
}

resource "aws_resiliencehubv2_service" "test" {
  name       = var.rName
  policy_arn = aws_resiliencehubv2_policy.test.arn
  regions    = [data.aws_region.current.name]

  permission_model {
    invoker_role_name = aws_iam_role.test.name
  }

  depends_on = [aws_iam_role_policy_attachment.service_AWSResilienceHubV2AssessmentExecutionPolicy]
}

resource "aws_resiliencehubv2_policy" "test" {
  name = var.policy_name

  availability_slo {
    target = 99.9
  }
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = "${var.rName}-invoker"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "resiliencehub.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "service_AWSResilienceHubV2AssessmentExecutionPolicy" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AWSResilienceHubV2AssessmentExecutionPolicy"
  role       = aws_iam_role.test.name
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "description" {
  type     = string
  nullable = false
}
