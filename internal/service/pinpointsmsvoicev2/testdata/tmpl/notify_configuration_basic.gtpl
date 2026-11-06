resource "aws_pinpointsmsvoicev2_notify_configuration" "test" {
{{- template "region" }}
  display_name     = var.rName
  use_case         = "CODE_VERIFICATION"
  enabled_channels = ["SMS"]

{{- template "tags" . }}
}
