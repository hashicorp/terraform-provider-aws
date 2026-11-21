resource "aws_cognito_user_pool_session_configuration" "test" {
  {{- template "region" }}
  user_pool_id          = aws_cognito_user_pool.test.id
  user_session_tracking = "ENABLED"

  token_session {
    max_per_user = 20
  }
}

resource "aws_cognito_user_pool" "test" {
  {{- template "region" }}
  name           = var.rName
  user_pool_tier = "ESSENTIALS"
}
