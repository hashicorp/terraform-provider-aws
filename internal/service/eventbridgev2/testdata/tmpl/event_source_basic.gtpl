resource "aws_eventbridgev2_event_source" "test" {
{{- template "region" }}
  name          = var.rName
  event_bus_arn = aws_eventbridgev2_event_bus.test.arn

  configuration {
    aws_service_events_configuration {
      aws_service = "aws.s3"
    }
  }
{{- template "tags" . }}
}

resource "aws_eventbridgev2_event_bus" "test" {
{{- template "region" }}
  name = var.rName
}
