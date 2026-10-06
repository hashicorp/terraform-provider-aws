# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_resource_policy" "test" {
  count = var.resource_count

  resource_arn = aws_bedrockagentcore_agent_runtime.test[count.index].agent_runtime_arn
  policy       = data.aws_iam_policy_document.resource_policy[count.index].json
}

data "aws_iam_policy_document" "resource_policy" {
  count = var.resource_count

  statement {
    effect = "Allow"
    actions = [
      "bedrock-agentcore:InvokeAgentRuntime",
    ]
    principals {
      type        = "*"
      identifiers = ["*"]
    }
    resources = [
      aws_bedrockagentcore_agent_runtime.test[count.index].agent_runtime_arn
    ]
  }
}

resource "aws_bedrockagentcore_agent_runtime" "test" {
  count = var.resource_count

  agent_runtime_name = "${var.rName}_${count.index}"
  role_arn           = aws_iam_role.test.arn

  agent_runtime_artifact {
    code_configuration {
      entry_point = ["runtime_example.py"]
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

resource "aws_iam_role" "test" {
  name               = var.rName
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
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

resource "aws_s3_bucket" "test" {
  bucket        = replace(var.rName, "_", "-")
  force_destroy = true
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = "runtime_example.zip"
  source = "test-fixtures/runtime_example.zip"
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
