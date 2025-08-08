resource "aws_inspector_assessment_target" "test" {
{{- template "region" }}
  name = var.rName
{{- template "tags" }}
}
