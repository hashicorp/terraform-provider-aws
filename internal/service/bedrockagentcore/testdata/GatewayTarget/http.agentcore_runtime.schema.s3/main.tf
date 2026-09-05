# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_gateway_target" "test" {
  name               = var.rName
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id

  credential_provider_configuration {
    gateway_iam_role {}
  }

  target_configuration {
    http {
      agentcore_runtime {
        arn = aws_bedrockagentcore_agent_runtime.test.agent_runtime_arn

        schema {
          source {
            s3 {
              uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.test.key}"
            }
          }
        }
      }
    }
  }
}

resource "aws_bedrockagentcore_gateway" "test" {
  name            = var.rName
  role_arn        = aws_iam_role.test.arn
  authorizer_type = "AWS_IAM"
}

resource "aws_bedrockagentcore_agent_runtime" "test" {
  agent_runtime_name = replace(var.rName, "-", "_")
  role_arn           = aws_iam_role.test.arn

  agent_runtime_artifact {
    container_configuration {
      container_uri = var.container_uri
    }
  }

  network_configuration {
    network_mode = "PUBLIC"
  }
}

data "aws_iam_policy_document" "test_assume" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["bedrock-agentcore.amazonaws.com"]
    }
  }
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = [
      "ecr:GetAuthorizationToken",
      "ecr:BatchGetImage",
      "ecr:GetDownloadUrlForLayer"
    ]
    effect    = "Allow"
    resources = ["*"]
  }
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.id
  policy = data.aws_iam_policy_document.test.json
}

resource "aws_iam_role" "test" {
  name               = var.rName
  assume_role_policy = data.aws_iam_policy_document.test_assume.json
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = "schema.json"

  content = jsonencode({
    openapi = "3.0.0"
    info = {
      title   = "Test API"
      version = "1.0.0"
    }
    paths = {}
  })
}

resource "aws_s3_bucket" "test" {
  bucket        = var.rName
  force_destroy = true
}

data "aws_partition" "current" {}

resource "aws_iam_policy" "s3_bucket_full_access" {
  name        = "${var.rName}-s3-bucket"
  description = "Full S3 permissions for one bucket and its objects"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = "s3:*"
      Resource = [
        "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.bucket}",
        "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.bucket}/*",
      ]
    }]
  })
}

resource "aws_iam_role_policy_attachment" "s3_bucket_full_access" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.s3_bucket_full_access.arn
}

variable "container_uri" {
  type     = string
  nullable = false
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
