# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

# tflint-ignore: terraform_unused_declarations
data "aws_accountaccess_entitlements" "test" {
  application_arn = aws_accountaccess_application.test.arn

  filter {
    principal_role {
      account_id = split(":", aws_iam_role.test.arn)[4]
    }
  }

  depends_on = [aws_accountaccess_entitlement.test_user, aws_accountaccess_entitlement.test_group]
}

resource "aws_accountaccess_entitlement" "test_user" {
  application_arn = aws_accountaccess_application.test.arn

  entitlement {
    principal_role {
      role_arn = aws_iam_role.test.arn

      principal {
        identity_center {
          user_id = aws_identitystore_user.test.user_id
        }
      }
    }
  }
}

resource "aws_accountaccess_entitlement" "test_group" {
  application_arn = aws_accountaccess_application.test.arn

  entitlement {
    principal_role {
      role_arn = aws_iam_role.test.arn

      principal {
        identity_center {
          group_id = aws_identitystore_group.test.group_id
        }
      }
    }
  }
}

data "aws_ssoadmin_instances" "test" {}

locals {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
  instance_arn      = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}

resource "aws_accountaccess_application" "test" {
  identity_source {
    identity_center {
      instance_arn = local.instance_arn
    }
  }
}

resource "aws_identitystore_user" "test" {
  identity_store_id = local.identity_store_id
  display_name      = var.rName
  user_name         = var.rName

  name {
    given_name  = "Acceptance"
    family_name = "Test"
  }

  emails {
    value = "${var.rName}@example.com"
  }
}

resource "aws_identitystore_group" "test" {
  identity_store_id = local.identity_store_id
  display_name      = var.rName
  description       = "Account Access acceptance test group"
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

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
