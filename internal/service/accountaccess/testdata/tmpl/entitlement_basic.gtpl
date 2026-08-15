data "aws_ssoadmin_instances" "test" {}

locals {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
  instance_arn      = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}

resource "aws_identitystore_user" "test" {
  identity_store_id = local.identity_store_id

  display_name = var.rName
  user_name    = var.rName

  name {
    given_name  = "Acceptance"
    family_name = "Test"
  }

  emails {
    value = "${var.rName}@example.com"
  }
}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "account-access.amazonaws.com"
        }
        Action = [
          "sts:AssumeRole",
          "sts:SetContext",
          "sts:TagSession",
        ]
      },
    ]
  })
}

resource "aws_accountaccess_application" "test" {
  identity_center_instance_arn = local.instance_arn
}

resource "aws_accountaccess_entitlement" "test" {
  application_arn = aws_accountaccess_application.test.arn
  principal_id    = aws_identitystore_user.test.user_id
  principal_type  = "USER"
  role_arn        = aws_iam_role.test.arn
{{- template "region" }}
}
