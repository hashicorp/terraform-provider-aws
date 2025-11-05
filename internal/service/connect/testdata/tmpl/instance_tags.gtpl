resource "aws_connect_instance" "test" {
{{- template "region" }}
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = var.rName
  outbound_calls_enabled   = true
{{- template "tags" }}
}

