resource "aws_pinpointsmsvoicev2_phone_number" "test" {
{{- template "region" }}
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]

{{- template "tags" . }}
}
