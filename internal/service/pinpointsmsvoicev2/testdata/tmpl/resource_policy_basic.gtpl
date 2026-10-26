resource "aws_pinpointsmsvoicev2_resource_policy" "test" {
{{- template "region" }}
  resource_arn = aws_pinpointsmsvoicev2_phone_number.test.arn
  policy       = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    actions   = ["sms-voice:SendTextMessage"]
    resources = [aws_pinpointsmsvoicev2_phone_number.test.arn]
    principals {
      type        = "AWS"
      identifiers = [data.aws_caller_identity.current.account_id]
    }
  }
}

data "aws_caller_identity" "current" {}

resource "aws_pinpointsmsvoicev2_phone_number" "test" {
{{- template "region" }}
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}
