resource "aws_secretsmanager_secret_policy" "test" {
{{- template "region" }}
  secret_arn = aws_secretsmanager_secret.test.arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid    = "EnableAllPermissions"
      Effect = "Allow"
      Principal = {
        AWS = "*"
      }
      Action   = "secretsmanager:GetSecretValue"
      Resource = "*"
    }]
  })
}

resource "aws_secretsmanager_secret" "test" {
{{- template "region" }}
  name = var.rName
}

