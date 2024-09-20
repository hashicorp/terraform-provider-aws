resource "aws_apprunner_connection" "test" {
  connection_name = var.rName
  provider_type   = "GITHUB"

{{- template "tags" . }}
}
