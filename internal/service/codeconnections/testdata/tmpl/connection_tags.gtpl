resource "aws_codeconnections_connection" "test" {
  name          = var.rName
  provider_type = "Bitbucket"
{{- template "tags" . }}
}
