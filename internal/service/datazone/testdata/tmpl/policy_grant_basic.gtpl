resource "aws_datazone_policy_grant" "test" {
{{- template "region" }}
  domain_identifier = aws_datazone_domain.test.id
  entity_type       = "DOMAIN_UNIT"
  entity_identifier = aws_datazone_domain.test.root_domain_unit_id
  policy_type       = "CREATE_DOMAIN_UNIT"

  detail {
    create_domain_unit {}
  }

  principal {
    user {
      all_users_grant_filter {}
    }
  }
}

resource "aws_datazone_domain" "test" {
{{- template "region" }}
  name                  = var.rName
  domain_execution_role = aws_iam_role.domain_execution_role.arn
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
    ]
  })

  inline_policy {
    name = var.rName
    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Action   = ["datazone:*", "ram:*", "sso:*", "kms:*"]
          Effect   = "Allow"
          Resource = "*"
        },
      ]
    })
  }
}
