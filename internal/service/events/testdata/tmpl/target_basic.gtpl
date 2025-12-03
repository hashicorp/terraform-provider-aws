resource "aws_cloudwatch_event_target" "test" {
{{- template "region" }}
  rule      = aws_cloudwatch_event_rule.test.name
  target_id = var.rName
  arn       = aws_sns_topic.test.arn
}

resource "aws_cloudwatch_event_rule" "test" {
{{- template "region" }}
  name                = var.rName
  schedule_expression = "rate(1 hour)"
{{- template "tags" }}
}

resource "aws_sns_topic" "test" {
{{- template "region" }}
  name = var.rName
}
