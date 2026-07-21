resource "aws_pinpointsmsvoicev2_keyword" "test" {
{{- template "region" }}
  origination_identity = aws_pinpointsmsvoicev2_phone_number.test.id
  keyword              = var.rName
  keyword_message      = "test keyword message"
}

resource "aws_pinpointsmsvoicev2_phone_number" "test" {
{{- template "region" }}
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}
