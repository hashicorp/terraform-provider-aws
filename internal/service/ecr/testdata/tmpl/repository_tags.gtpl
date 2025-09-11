resource "aws_ecr_repository" "test" {
{{- template "region" }}
  name = var.rName
{{- template "tags" }}
}
