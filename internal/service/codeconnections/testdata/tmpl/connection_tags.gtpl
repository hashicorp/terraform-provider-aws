resource "aws_codeconnections_connection" "test" {
{{- template "region" }}
  name          = var.rName
  provider_type = "Bitbucket"
{{- template "tags" . }}
}
