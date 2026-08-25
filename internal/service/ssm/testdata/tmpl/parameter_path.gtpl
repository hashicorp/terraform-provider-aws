resource "aws_ssm_parameter" "test" {
{{- template "region" }}
  name  = "/path/${var.rName}"
  type  = "String"
  value = var.rName

{{- template "tags" . }}
}
