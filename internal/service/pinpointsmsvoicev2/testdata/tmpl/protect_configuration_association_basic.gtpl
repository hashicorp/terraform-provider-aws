resource "aws_pinpointsmsvoicev2_protect_configuration_association" "test" {
{{- template "region" }}
  configuration_set_name   = aws_pinpointsmsvoicev2_configuration_set.test.name
  protect_configuration_id = aws_pinpointsmsvoicev2_protect_configuration.test.id
}

resource "aws_pinpointsmsvoicev2_configuration_set" "test" {
{{- template "region" }}
  name = var.rName
}

resource "aws_pinpointsmsvoicev2_protect_configuration" "test" {
{{- template "region" }}
}
