resource "aws_bedrockagentcore_agent_runtime_endpoint" "test" {
{{- template "region" }}
  name                  = var.rName
  agent_runtime_id      = aws_bedrockagentcore_agent_runtime.test.agent_runtime_id
  agent_runtime_version = aws_bedrockagentcore_agent_runtime.test.agent_runtime_version

{{- template "tags" . }}
}

resource "aws_bedrockagentcore_agent_runtime" "test" {
{{- template "region" }}
  agent_runtime_name = var.rName
  role_arn           = aws_iam_role.test.arn

  agent_runtime_artifact {
    container_configuration {
      container_uri = var.rImageUri
    }
  }

  network_configuration {
    network_mode = "PUBLIC"
  }

  depends_on = [aws_iam_role_policy.test]
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

data "aws_iam_policy_document" "test" {
  statement {
    actions = [
      "ecr:GetAuthorizationToken",
      "ecr:BatchGetImage",
      "ecr:GetDownloadUrlForLayer",
    ]
    effect    = "Allow"
    resources = ["*"]
  }
}

resource "aws_iam_role" "test" {
  name               = var.rName
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy" "test" {
  name   = var.rName
  role   = aws_iam_role.test.id
  policy = data.aws_iam_policy_document.test.json
}
