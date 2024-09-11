resource "aws_appconfig_environment" "test" {
  name           = var.rName
  application_id = aws_appconfig_application.test.id

{{- template "tags" . }}
}

resource "aws_appconfig_application" "test" {
  name = var.rName
}
