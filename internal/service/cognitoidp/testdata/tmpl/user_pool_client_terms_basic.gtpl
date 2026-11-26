variable "terms_name" {
  type    = string
  default = "terms-of-use"
}

variable "links" {
  type = map(string)
  default = {
    "cognito:default" = "https://example.com/terms"
    "cognito:french"  = "https://example.com/fr/terms"
  }
}

resource "aws_cognito_user_pool" "test" {
  {{- template "region" }}
  name           = var.rName
  user_pool_tier = "ESSENTIALS"
}

resource "aws_cognito_user_pool_client" "test" {
  {{- template "region" }}
  name                                 = var.rName
  user_pool_id                         = aws_cognito_user_pool.test.id
  allowed_oauth_flows_user_pool_client = true
  allowed_oauth_flows                  = ["code"]
  allowed_oauth_scopes                 = ["openid"]
  callback_urls                        = ["https://example.com/callback"]
  supported_identity_providers         = ["COGNITO"]
}

resource "aws_cognito_user_pool_domain" "test" {
  {{- template "region" }}
  domain                = var.rName
  user_pool_id          = aws_cognito_user_pool.test.id
  managed_login_version = 2
}

resource "aws_cognito_user_pool_client_terms" "test" {
  {{- template "region" }}
  client_id    = aws_cognito_user_pool_client.test.id
  links        = var.links
  terms_name   = var.terms_name
  user_pool_id = aws_cognito_user_pool.test.id

  depends_on = [aws_cognito_user_pool_domain.test]
}
