resource "aws_db_parameter_group" "test" {
{{- template "region" }}
  name   = var.rName
  family = "mysql5.6"
{{- template "tags" . }}
}
