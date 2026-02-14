resource "aws_sesv2_tenant" "test" {
  tenant_name = var.rName
{{- template "tags" . }}
}