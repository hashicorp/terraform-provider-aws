resource "aws_iot_billing_group" "test" {
{{- template "region" }}
  name = var.rName
{{- template "tags" . }}
}
