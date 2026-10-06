# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_agent_runtime_endpoint" "test" {
  name                  = var.rName
  agent_runtime_id      = aws_bedrockagentcore_agent_runtime.test.agent_runtime_id
  agent_runtime_version = aws_bedrockagentcore_agent_runtime.test.agent_runtime_version

  tags = var.resource_tags
}

# testAccAgentRuntimeConfig_codeConfiguration

resource "aws_bedrockagentcore_agent_runtime" "test" {
  agent_runtime_name = var.rName
  role_arn           = aws_iam_role.test.arn

  agent_runtime_artifact {
    code_configuration {
      entry_point = ["main.py"]
      runtime     = "PYTHON_3_13"
      code {
        s3 {
          bucket = aws_s3_bucket.test.bucket
          prefix = aws_s3_object.test.key
        }
      }
    }
  }

  network_configuration {
    network_mode = "PUBLIC"
  }

  depends_on = [aws_iam_role_policy.bucket]

  tags = var.resource_tags
}

# testAccAgentRuntimeConfig_baseIAMRole

resource "aws_iam_role" "test" {
  name               = var.rName
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["bedrock-agentcore.amazonaws.com"]
    }
  }
}

# testAccAgentRuntimeConfig_baseS3Bucket

resource "aws_s3_bucket" "test" {
  bucket        = replace(var.rName, "_", "-")
  force_destroy = true
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = "agent-runtime-codezip.zip"
  source = "${path.module}/test-fixtures/agent-runtime-codezip.zip"
}

resource "aws_iam_role_policy" "bucket" {
  name   = var.rName
  role   = aws_iam_role.test.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = ["s3:GetObject", "s3:GetObjectVersion"]
      Resource = "${aws_s3_bucket.test.arn}/*"
    }]
  })
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}
