# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

data "aws_region" "current" {
}

resource "aws_resiliencehubv2_service" "test" {
  name    = var.rName
  regions = [data.aws_region.current.name]

  permission_model {
    invoker_role_name = aws_iam_role.test.name
  }

  associated_system {
    system_arn       = aws_resiliencehubv2_system.test.arn
    user_journey_ids = var.user_journey_count > 0 ? slice(aws_resiliencehubv2_user_journey.test[*].user_journey_id, 0, var.user_journey_count) : null
  }

  depends_on = [aws_iam_role_policy_attachment.service_AWSResilienceHubV2AssessmentExecutionPolicy]
}

resource "aws_resiliencehubv2_user_journey" "test" {
  count = 2

  name       = "${var.rName}-${count.index}"
  system_arn = aws_resiliencehubv2_system.test.arn
}

resource "aws_resiliencehubv2_system" "test" {
  name = var.rName
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

variable "user_journey_count" {
  type     = number
  nullable = false
  validation {
    condition     = var.user_journey_count >= 0 && var.user_journey_count <= 2
    error_message = "Value must be between 0 and 2 (inclusive)."
  }
}
