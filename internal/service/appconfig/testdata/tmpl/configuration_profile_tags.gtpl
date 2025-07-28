resource "aws_appconfig_configuration_profile" "test" {
  application_id = aws_appconfig_application.test.id
  location_uri   = "hosted"
  name           = var.rName

{{- template "tags" . }}
}

resource "aws_appconfig_application" "test" {
  name = var.rName
}
