resource "aws_bedrockagentcore_gateway_rate_limit" "test" {
{{- template "region" }}
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id
  dimension_keys     = ["targetName"]

  entries {
    dimensions = {
      targetName = "*"
    }

    requests {
      rate   = 100
      period = "second"
    }
  }
}

data "aws_iam_policy_document" "gateway_assume" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["bedrock-agentcore.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "gateway" {
  name               = "${var.rName}-gateway"
  assume_role_policy = data.aws_iam_policy_document.gateway_assume.json
}

resource "aws_bedrockagentcore_gateway" "test" {
{{- template "region" }}
  name     = "${var.rName}-gateway"
  role_arn = aws_iam_role.gateway.arn

  authorizer_type = "CUSTOM_JWT"
  authorizer_configuration {
    custom_jwt_authorizer {
      discovery_url    = "https://accounts.google.com/.well-known/openid-configuration"
      allowed_audience = ["test"]
    }
  }
}
