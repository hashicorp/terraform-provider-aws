data "aws_appconfig_environment" "test" {
  application_id = aws_appconfig_application.test.id
  environment_id = aws_appconfig_environment.test.environment_id
}
