resource "aws_cloudwatch_event_rule" "test" {
{{- template "region" }}
  name                = var.rName
  schedule_expression = "rate(1 hour)"
{{- template "tags" }}
}
