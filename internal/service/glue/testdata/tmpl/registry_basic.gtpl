resource "aws_glue_registry" "test" {
{{- template "region" }}
  registry_name = var.rName
}
