resource "aws_ssm_parameter" "test" {
  name  = var.rName
  type  = "String"
  value = var.rName

{{- template "tags" . }}
}
