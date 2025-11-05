resource "aws_secretsmanager_secret_version" "test" {
{{- template "region" }}
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = "test-string"
}

resource "aws_secretsmanager_secret" "test" {
{{- template "region" }}
  name = var.rName
}
