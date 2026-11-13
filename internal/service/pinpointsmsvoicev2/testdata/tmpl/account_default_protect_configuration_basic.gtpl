resource "aws_pinpointsmsvoicev2_account_default_protect_configuration" "test" {
{{- template "region" }}
  protect_configuration_id = aws_pinpointsmsvoicev2_protect_configuration.test.id
}

resource "aws_pinpointsmsvoicev2_protect_configuration" "test" {
{{- template "region" }}
}
