resource "aws_appconfig_deployment" "test"{
  application_id           = aws_appconfig_application.test.id
  configuration_profile_id = aws_appconfig_configuration_profile.test.configuration_profile_id
  configuration_version    = aws_appconfig_hosted_configuration_version.test.version_number
  deployment_strategy_id   = aws_appconfig_deployment_strategy.test.id
  environment_id           = aws_appconfig_environment.test.environment_id

{{- template "tags" . }}
}

resource "aws_appconfig_application" "test" {
  name = var.rName
}

resource "aws_appconfig_environment" "test" {
  name           = var.rName
  application_id = aws_appconfig_application.test.id
}

resource "aws_appconfig_configuration_profile" "test" {
  application_id = aws_appconfig_application.test.id
  name           = var.rName
  location_uri   = "hosted"
}

resource "aws_appconfig_deployment_strategy" "test" {
  name = var.rName

  deployment_duration_in_minutes = 3
  growth_factor                  = 10
  replicate_to                   = "NONE"
}

resource "aws_appconfig_hosted_configuration_version" "test" {
  application_id           = aws_appconfig_application.test.id
  configuration_profile_id = aws_appconfig_configuration_profile.test.configuration_profile_id
  content_type             = "application/json"

  content = jsonencode({
    foo = "bar"
  })

  description = var.rName
}
