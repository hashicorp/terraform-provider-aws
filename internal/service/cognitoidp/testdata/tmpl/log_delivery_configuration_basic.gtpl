resource "aws_cognito_log_delivery_configuration" "test" {
  {{- template "region" }}
  user_pool_id = aws_cognito_user_pool.test.id

  log_configurations {
    event_source = "userNotification"
    log_level    = "ERROR"

    cloud_watch_logs_configuration {
      log_group_arn = aws_cloudwatch_log_group.test.arn
    }
  }
}

resource "aws_cognito_user_pool" "test" {
  {{- template "region" }}
  name = var.rName
}

resource "aws_cloudwatch_log_group" "test" {
  {{- template "region" }}
  name = var.rName
}
