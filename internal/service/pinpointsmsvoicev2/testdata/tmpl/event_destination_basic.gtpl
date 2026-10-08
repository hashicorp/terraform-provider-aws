resource "aws_pinpointsmsvoicev2_event_destination" "test" {
{{- template "region" }}
  configuration_set_name = aws_pinpointsmsvoicev2_configuration_set.test.name
  event_destination_name = var.rName

  matching_event_types = ["TEXT_DELIVERED"]

  sns_destination {
    topic_arn = aws_sns_topic.test.arn
  }
}

resource "aws_pinpointsmsvoicev2_configuration_set" "test" {
{{- template "region" }}
  name = var.rName
}

resource "aws_sns_topic" "test" {
{{- template "region" }}
  name = var.rName
}
