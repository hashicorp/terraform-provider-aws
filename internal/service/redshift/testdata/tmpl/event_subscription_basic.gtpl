resource "aws_redshift_event_subscription" "test" {
{{- template "region" }}
  name          = var.rName
  sns_topic_arn = aws_sns_topic.test.arn
{{- template "tags" . }}
}

resource "aws_sns_topic" "test" {
{{- template "region" }}
  name = var.rName
}
