resource "aws_sfn_activity" "test" {
{{- template "region" }}
  name = var.rName
{{- template "tags" }}
}