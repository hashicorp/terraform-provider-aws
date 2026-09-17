resource "aws_bedrockagentcore_resource_policy" "test" {
{{- template "region" }}
  resource_arn = aws_bedrockagentcore_agent_runtime.test.agent_runtime_arn
  policy       = data.aws_iam_policy_document.resource_policy.json
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

resource "aws_iam_role" "test" {
  name               = var.rName
  assume_role_policy = data.aws_iam_policy_document.test_assume.json
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.id
  policy = data.aws_iam_policy_document.test.json
}

resource "aws_bedrockagentcore_agent_runtime" "test" {
{{- template "region" }}
  agent_runtime_name = var.rName
  role_arn           = aws_iam_role.test.arn

  agent_runtime_artifact {
    container_configuration {
      container_uri = var.AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V1_URI
    }
  }

  network_configuration {
    network_mode = "PUBLIC"
  }
}

data "aws_iam_policy_document" "resource_policy" {
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
      aws_bedrockagentcore_agent_runtime.test.agent_runtime_arn
    ]
  }
}
