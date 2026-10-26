# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_accountaccess_entitlement" "test" {
  provider = aws

  include_resource = true

  config {
    application_arn = aws_accountaccess_application.test.arn

    filter {
      principal_role {
        principal {
          identity_center {
            user_id = aws_identitystore_user.test.user_id
          }
        }
      }
    }
  }
}
