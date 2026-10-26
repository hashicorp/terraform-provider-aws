resource "aws_bedrockagentcore_gateway" "test" {
{{- template "region" }}
  name     = var.rName
  role_arn = aws_iam_role.test.arn

  authorizer_type = "CUSTOM_JWT"

  authorizer_configuration {
    custom_jwt_authorizer {
      discovery_url    = "https://accounts.google.com/.well-known/openid-configuration"
      allowed_audience = ["test1", "test2"]
    }
  }

  protocol_type = "MCP"

  depends_on = [aws_iam_role_policy_attachment.test]

{{- template "tags" . }}
}

data "aws_partition" "current" {}

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

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/BedrockAgentCoreFullAccess"
}
