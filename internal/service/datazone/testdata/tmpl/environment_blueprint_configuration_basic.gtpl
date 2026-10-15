resource "aws_datazone_environment_blueprint_configuration" "test" {
{{- template "region" }}
  domain_id                = aws_datazone_domain.test.id
  environment_blueprint_id = data.aws_datazone_environment_blueprint.test.id
  enabled_regions          = []
}

resource "aws_iam_role" "test" {
  name = var.rName
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = ["sts:AssumeRole", "sts:TagSession"]
        Effect = "Allow"
        Principal = {
          Service = "datazone.amazonaws.com"
        }
      },
      {
        Action = ["sts:AssumeRole", "sts:TagSession"]
        Effect = "Allow"
        Principal = {
          Service = "cloudformation.amazonaws.com"
        }
      },
    ]
  })

  inline_policy {
    name = var.rName
    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Action = [
            "datazone:*",
            "ram:*",
            "sso:*",
            "kms:*",
          ]
          Effect   = "Allow"
          Resource = "*"
        },
      ]
    })
  }
}

resource "aws_datazone_domain" "test" {
{{- template "region" }}
  name                  = var.rName
  domain_execution_role = aws_iam_role.test.arn
}

data "aws_datazone_environment_blueprint" "test" {
{{- template "region" }}
  domain_id = aws_datazone_domain.test.id
  name      = "DefaultDataLake"
  managed   = true
}
