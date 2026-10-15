resource "aws_pinpointsmsvoicev2_pool" "test" {
{{- template "region" }}
  iso_country_code       = "US"
  message_type           = "TRANSACTIONAL"
  origination_identities = [aws_pinpointsmsvoicev2_phone_number.test.arn]

{{- template "tags" . }}
}

resource "aws_pinpointsmsvoicev2_phone_number" "test" {
{{- template "region" }}
  force_disassociate  = true
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}
