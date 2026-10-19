resource "aws_dms_instance_profile" "test" {
{{- template "region" }}
{{- template "tags" . }}
}
