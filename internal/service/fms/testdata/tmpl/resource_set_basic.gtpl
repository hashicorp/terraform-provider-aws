data "aws_caller_identity" "current" {}

resource "aws_fms_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id
}

resource "aws_fms_resource_set" "test" {
  depends_on = [aws_fms_admin_account.test]
  resource_set {
    name               = var.rName
    resource_type_list = ["AWS::NetworkFirewall::Firewall"]
  }
{{- template "tags" . }}
}
