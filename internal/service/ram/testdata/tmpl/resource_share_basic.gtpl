resource "aws_ram_resource_share" "test" {
{{- template "region" }}
  name = var.rName
{{- template "tags" . }}
}
