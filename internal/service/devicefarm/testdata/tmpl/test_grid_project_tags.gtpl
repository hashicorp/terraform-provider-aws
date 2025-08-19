resource "aws_devicefarm_test_grid_project" "test" {
{{- template "region" }}
  name = var.rName
{{- template "tags" . }}
}
