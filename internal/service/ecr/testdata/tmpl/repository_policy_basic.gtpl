resource "aws_ecr_repository_policy" "test" {
{{- template "region" }}
  repository = aws_ecr_repository.test.name

  policy = jsonencode({
    Version = "2008-10-17"
    Statement = [{
      Sid       = var.rName
      Effect    = "Allow"
      Principal = "*"
      Action    = "ecr:ListImages"
    }]
  })
}

resource "aws_ecr_repository" "test" {
{{- template "region" }}
  name = var.rName
}
