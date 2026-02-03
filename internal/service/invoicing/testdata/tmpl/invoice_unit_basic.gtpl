resource "aws_invoicing_invoice_unit" "test" {
{{- template "region" }}
  name             = var.rName
  invoice_receiver = data.aws_caller_identity.current.account_id

  rule {
    linked_accounts = [data.aws_caller_identity.current.account_id]
  }

{{- template "tags" . }}
}

data "aws_caller_identity" "current" {}
