resource "aws_eventbridgev2_event_bus" "test" {
{{- template "region" }}
  name = var.rName
{{- template "tags" . }}
}
