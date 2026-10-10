resource "aws_pinpointsmsvoicev2_configuration_set" "test" {
{{- template "region" }}
  name = var.rName

{{- template "tags" . }}
}
