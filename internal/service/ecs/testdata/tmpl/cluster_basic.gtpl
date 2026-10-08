resource "aws_ecs_cluster" "test" {
{{- template "region" }}
  name = var.rName
{{- template "tags" . }}
}
