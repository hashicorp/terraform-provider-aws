resource "aws_appintegrations_event_integration" "test" {
  name            = var.rName
  eventbridge_bus = "default"

  event_filter {
    source = "aws.partner/examplepartner.com"
  }

{{- template "tags" . }}
}
