resource "aws_devicefarm_instance_profile" "test" {
{{- template "region" }}
  name = var.rName
{{- template "tags" . }}
}
