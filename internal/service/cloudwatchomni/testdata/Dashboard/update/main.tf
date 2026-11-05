# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatchomni_dashboard" "test" {
  space_id    = aws_cloudwatchomni_space.test.space_id
  name        = var.dashboard_name
  description = var.description
  body = jsonencode({
    panels = [for content in var.markdown_contents : {
      type   = "markdown"
      layout = { x = 0, y = 0, w = 60, h = "auto" }
      config = { content = content }
    }]
  })
}

resource "aws_cloudwatchomni_space" "test" {
  name                 = var.rName
  domain_id            = aws_cloudwatchomni_domain.test.domain_id
  data_access_role_arn = aws_iam_role.test.arn

  depends_on = [aws_iam_role_policy_attachment.test]
}

resource "aws_cloudwatchomni_domain" "test" {
  name               = var.rName
  identity_providers = ["IAM"]
}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "cloudwatch.${data.aws_partition.current.dns_suffix}" }
      Condition = { StringEquals = { "aws:SourceAccount" = data.aws_caller_identity.current.account_id } }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/CloudWatchOmniSpaceAccessPolicy"
}

data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "dashboard_name" {
  description = "Name of the dashboard"
  type        = string
  nullable    = false
}

variable "description" {
  description = "Description of the dashboard"
  type        = string
  nullable    = true
  default     = null
}

variable "markdown_contents" {
  description = "Contents of the dashboard's markdown panels"
  type        = list(string)
  nullable    = false
  default     = []
}
