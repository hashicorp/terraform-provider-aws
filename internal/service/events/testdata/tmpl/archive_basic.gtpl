resource "aws_cloudwatch_event_archive" "test" {
{{- template "region" }}
  name             = var.rName
  event_source_arn = aws_cloudwatch_event_bus.test.arn
}

resource "aws_cloudwatch_event_bus" "test" {
{{- template "region" }}
  name = "${var.rName}-bus"
}
