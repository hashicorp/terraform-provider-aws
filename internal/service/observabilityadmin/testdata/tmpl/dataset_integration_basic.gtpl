resource "aws_observabilityadmin_dataset_integration" "test" {
{{- template "region" }}
  role_arn = aws_iam_role.test.arn

{{- template "tags" . }}
}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "logs.amazonaws.com"
      }
    }]
  })
}
