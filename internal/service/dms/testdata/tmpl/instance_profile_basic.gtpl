resource "aws_dms_instance_profile" "test" {
{{- template "region" }}
  name = var.rName
{{- template "tags" . }}
}
