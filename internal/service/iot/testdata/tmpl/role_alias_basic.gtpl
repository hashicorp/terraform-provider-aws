resource "aws_iam_role" "role" {
  name = var.rName

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = {
      Effect = "Allow"
      Principal = {
        Service = "credentials.iot.amazonaws.com"
      }
      Action = "sts:AssumeRole"
    }
  })
}

resource "aws_iot_role_alias" "test" {
{{- template "region" }}
  alias    = var.rName
  role_arn = aws_iam_role.role.arn
{{- template "tags" . }}
}
