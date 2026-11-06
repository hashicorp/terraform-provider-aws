resource "aws_bedrockagentcore_consent_portal" "test" {
{{- template "region" }}
  name               = var.rName
  execution_role_arn = aws_iam_role.test.arn

  idp_config {
    credential_provider_arn = aws_bedrockagentcore_oauth2_credential_provider.test.credential_provider_arn
    scopes                  = ["openid"]
  }

  sources {
    identifier = aws_bedrockagentcore_gateway.test.gateway_id
    type       = "agentcore-gateway"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
{{- template "tags" . }}
}

resource "aws_cognito_user_pool" "test" {
{{- template "region" }}
  name = var.rName
}

resource "aws_cognito_user_pool_domain" "test" {
{{- template "region" }}
  domain       = var.rName
  user_pool_id = aws_cognito_user_pool.test.id
}

resource "aws_cognito_user_pool_client" "test" {
{{- template "region" }}
  name                                 = var.rName
  user_pool_id                         = aws_cognito_user_pool.test.id
  generate_secret                      = true
  allowed_oauth_flows                  = ["code"]
  allowed_oauth_flows_user_pool_client = true
  allowed_oauth_scopes                 = ["openid", "email", "profile"]
  callback_urls                        = ["https://example.com/callback"]
  supported_identity_providers         = ["COGNITO"]
}

resource "aws_bedrockagentcore_oauth2_credential_provider" "test" {
{{- template "region" }}
  name                       = var.rName
  credential_provider_vendor = "CustomOauth2"
  depends_on                 = [aws_cognito_user_pool_domain.test]

  oauth2_provider_config {
    custom_oauth2_provider_config {
      client_id     = aws_cognito_user_pool_client.test.id
      client_secret = aws_cognito_user_pool_client.test.client_secret
      oauth_discovery {
        discovery_url = "https://${aws_cognito_user_pool.test.endpoint}/.well-known/openid-configuration"
      }
    }
  }
}

resource "aws_bedrockagentcore_gateway" "test" {
{{- template "region" }}
  name            = var.rName
  role_arn        = aws_iam_role.test.arn
  authorizer_type = "CUSTOM_JWT"
  protocol_type   = "MCP"

  authorizer_configuration {
    custom_jwt_authorizer {
      discovery_url   = "https://${aws_cognito_user_pool.test.endpoint}/.well-known/openid-configuration"
      allowed_clients = [aws_cognito_user_pool_client.test.id]
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}

data "aws_partition" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["bedrock-agentcore.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = var.rName
  assume_role_policy = data.aws_iam_policy_document.test.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/BedrockAgentCoreFullAccess"
}
