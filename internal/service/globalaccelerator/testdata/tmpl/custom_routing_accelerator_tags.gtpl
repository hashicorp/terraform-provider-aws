resource "aws_globalaccelerator_custom_routing_accelerator" "test" {
{{- template "region" }}
  name = var.rName
{{- template "tags" . }}
}
