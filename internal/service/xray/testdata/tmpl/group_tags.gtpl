resource "aws_xray_group" "test" {
  group_name        = var.rName
  filter_expression = "responsetime > 5"
{{- template "tags" . }}
}
