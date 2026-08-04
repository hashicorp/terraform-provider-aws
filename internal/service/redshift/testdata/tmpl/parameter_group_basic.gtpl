resource "aws_redshift_parameter_group" "test" {
{{- template "region" }}
  name   = var.rName
  family = "redshift-1.0"
{{- template "tags" . }}
}
