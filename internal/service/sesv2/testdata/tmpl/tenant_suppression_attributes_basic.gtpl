resource "aws_sesv2_tenant" "test" {
{{- template "region" }}
  tenant_name = var.rName
}

resource "aws_sesv2_tenant_suppression_attributes" "test" {
{{- template "region" }}
  tenant_name        = aws_sesv2_tenant.test.tenant_name
  suppression_scope  = "TENANT"
  suppressed_reasons = ["BOUNCE", "COMPLAINT"]
}
