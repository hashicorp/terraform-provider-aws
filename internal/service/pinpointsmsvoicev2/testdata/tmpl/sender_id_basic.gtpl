resource "aws_pinpointsmsvoicev2_sender_id" "test" {
{{- template "region" }}
  sender_id        = var.rName
  iso_country_code = "GB"

{{- template "tags" . }}
}
