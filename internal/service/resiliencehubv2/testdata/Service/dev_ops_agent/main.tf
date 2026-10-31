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

  # When agent_space_arn is empty, the block is omitted entirely, which removes
  # the DevOps Agent linkage (disconnect).
  dynamic "dev_ops_agent_configuration" {
    for_each = var.agent_space_arn == "" ? [] : [1]
    content {
      agent_space_arn = var.agent_space_arn
    }
  }

  depends_on = [
    aws_iam_role_policy_attachment.service_AWSResilienceHubV2AssessmentExecutionPolicy,
    aws_iam_role_policy.service_devops_agent,
  ]
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name               = "${var.rName}-invoker"
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

# Linking a DevOps Agent Agent Space requires the invoker role to be authorized to
# call DevOps Agent (aidevops). Resilience Hub assumes this role to validate the
# Agent Space (GetOperatorApp) and finalize the connection (ListAssociations).
resource "aws_iam_role_policy" "service_devops_agent" {
  count = var.agent_space_arn == "" ? 0 : 1
  name  = "${var.rName}-devops-agent"
  role  = aws_iam_role.test.name
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "aidevops:GetOperatorApp",
          "aidevops:ListAssociations",
        ]
        Resource = var.agent_space_arn
      },
    ]
  })
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "agent_space_arn" {
  description = "DevOps Agent Agent Space ARN to link; empty string removes the block (disconnect)"
  type        = string
  nullable    = false
  default     = ""
}
