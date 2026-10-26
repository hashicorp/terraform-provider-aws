resource "aws_cloudwatchomni_access_profile" "test" {
{{- template "region" }}
  name     = var.rName
  space_id = aws_cloudwatchomni_space.test.space_id
}

resource "aws_cloudwatchomni_space" "test" {
{{- template "region" }}
  name                 = var.rName
  domain_id            = aws_cloudwatchomni_domain.test.domain_id
  data_access_role_arn = aws_iam_role.test.arn
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "cloudwatch.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/CloudWatchAgentServerPolicy"
}

resource "aws_cloudwatchomni_domain" "test" {
{{- template "region" }}
  name               = var.rName
  identity_providers = ["IAM"]
}
