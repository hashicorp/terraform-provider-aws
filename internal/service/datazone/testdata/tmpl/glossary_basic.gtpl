resource "aws_datazone_glossary" "test" {
{{- template "region" }}
  description               = "desc"
  name                      = var.rName
  owning_project_identifier = aws_datazone_project.test.id
  status                    = "ENABLED"
  domain_identifier         = aws_datazone_domain.test.id
}

resource "aws_iam_role" "domain_execution_role" {
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
  domain_execution_role = aws_iam_role.domain_execution_role.arn
}

resource "aws_datazone_project" "test" {
{{- template "region" }}
  domain_identifier   = aws_datazone_domain.test.id
  name                = var.rName
  skip_deletion_check = true
}
