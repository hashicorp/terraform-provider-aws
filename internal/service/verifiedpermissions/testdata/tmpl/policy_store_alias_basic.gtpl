resource "aws_verifiedpermissions_policy_store" "test" {
{{- template "region" }}
  description = var.rName

  validation_settings {
    mode = "OFF"
  }
}

resource "aws_verifiedpermissions_policy_store_alias" "test" {
{{- template "region" }}
  alias_name      = "policy-store-alias/${var.rName}"
  policy_store_id = aws_verifiedpermissions_policy_store.test.policy_store_id
}