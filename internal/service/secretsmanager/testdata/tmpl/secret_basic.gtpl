resource "aws_secretsmanager_secret" "test" {
{{- template "region" }}
  name = var.rName

{{- template "tags" . }}
}
