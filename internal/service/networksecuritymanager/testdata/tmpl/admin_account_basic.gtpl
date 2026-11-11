resource "aws_networksecuritymanager_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id
  priority   = 1
}

data "aws_caller_identity" "current" {}
