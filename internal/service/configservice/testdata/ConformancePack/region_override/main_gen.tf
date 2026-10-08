# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_config_conformance_pack" "test" {
  region = var.region

  depends_on    = [aws_config_configuration_recorder.test]
  name          = var.rName
  template_body = <<EOT
Resources:
  IAMPasswordPolicy:
    Properties:
      ConfigRuleName: IAMPasswordPolicy
      Source:
        Owner: AWS
        SourceIdentifier: IAM_PASSWORD_POLICY
    Type: AWS::Config::ConfigRule
EOT
}

data "aws_partition" "current" {}

resource "aws_config_configuration_recorder" "test" {
  region = var.region

  depends_on = [aws_iam_role_policy_attachment.test]
  name       = var.rName
  role_arn   = aws_iam_role.test.arn
}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "config.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWS_ConfigRole"
  role       = aws_iam_role.test.name
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
