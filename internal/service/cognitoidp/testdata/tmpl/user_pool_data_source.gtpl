data "aws_cognito_user_pool" "test" {
  user_pool_id = aws_cognito_user_pool.test.id
}
