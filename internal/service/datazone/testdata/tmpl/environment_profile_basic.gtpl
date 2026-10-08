resource "aws_datazone_environment_profile" "test" {
{{- template "region" }}
  aws_account_id                   = data.aws_caller_identity.test.account_id
  aws_account_region               = data.aws_region.test.region
  environment_blueprint_identifier = data.aws_datazone_environment_blueprint.test.id
  name                             = var.rName
  project_identifier               = aws_datazone_project.test.id
  domain_identifier                = aws_datazone_domain.test.id
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

data "aws_caller_identity" "test" {}
data "aws_region" "test" {}

data "aws_datazone_environment_blueprint" "test" {
{{- template "region" }}
  domain_id = aws_datazone_domain.test.id
  name      = "DefaultDataLake"
  managed   = true
}

resource "aws_datazone_environment_blueprint_configuration" "test" {
{{- template "region" }}
  domain_id                = aws_datazone_domain.test.id
  environment_blueprint_id = data.aws_datazone_environment_blueprint.test.id
  provisioning_role_arn    = aws_iam_role.domain_execution_role.arn
  enabled_regions          = [data.aws_region.test.region]
}
