resource "aws_secretsmanager_secret" "test" {
  name = var.rName

{{- template "tags" . }}
}
