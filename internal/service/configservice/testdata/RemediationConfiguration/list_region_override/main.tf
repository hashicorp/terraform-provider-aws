# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_config_remediation_configuration" "test" {
  count  = var.resource_count
  region = var.region

  config_rule_name = aws_config_config_rule.test[count.index].name

  resource_type  = "AWS::S3::Bucket"
  target_id      = "AWS-EnableS3BucketEncryption"
  target_type    = "SSM_DOCUMENT"
  target_version = "1"

  parameter {
    name         = "AutomationAssumeRole"
    static_value = aws_iam_role.test.arn
  }
  parameter {
    name           = "BucketName"
    resource_value = "RESOURCE_ID"
  }
  parameter {
    name         = "SSEAlgorithm"
    static_value = "AES256"
  }
  automatic                  = false
  maximum_automatic_attempts = 2
  retry_attempt_seconds      = 30
  execution_controls {
    ssm_controls {
      concurrent_execution_rate_percentage = 75
      error_percentage                     = 10
    }
  }
}

resource "aws_config_config_rule" "test" {
  count  = var.resource_count
  region = var.region

  name = "${var.rName}-${count.index}"

  source {
    owner             = "AWS"
    source_identifier = "S3_BUCKET_VERSIONING_ENABLED"
  }

  depends_on = [aws_config_configuration_recorder.test]
}

# testAccConfigRuleConfig_base

data "aws_partition" "current" {}

resource "aws_config_configuration_recorder" "test" {
  region = var.region

  name     = var.rName
  role_arn = aws_iam_role.test.arn
}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "config.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
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

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
