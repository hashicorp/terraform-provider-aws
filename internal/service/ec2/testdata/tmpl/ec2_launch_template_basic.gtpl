resource "aws_launch_template" "test" {
{{- template "region" }}
  name = var.rName
}
