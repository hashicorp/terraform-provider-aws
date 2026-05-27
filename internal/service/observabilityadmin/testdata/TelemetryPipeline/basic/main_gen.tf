# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_observabilityadmin_telemetry_pipeline" "test" {

  name = var.rName

  configuration {
    body = yamlencode({
      pipeline = {
        source = {
          cloudwatch_logs = {
            aws = {
              sts_role_arn = aws_iam_role.test.arn
            }
            log_event_metadata = {
              data_source_name = replace(var.rName, "-", "_")
              data_source_type = "default"
            }
          }
        }
        sink = [{
          cloudwatch_logs = {
            log_group = "@original"
          }
        }]
      }
    })
  }

  depends_on = [aws_iam_role_policy.test]
}

# testAccTelemetryPipelineConfig_base

data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "observabilityadmin.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.name

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
      Resource = "arn:${data.aws_partition.current.partition}:logs:*:${data.aws_caller_identity.current.account_id}:*"
    }]
  })
}


variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
