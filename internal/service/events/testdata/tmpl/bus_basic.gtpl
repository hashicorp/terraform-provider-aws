resource "aws_cloudwatch_event_bus" "test" {
{{- template "region" }}
  name = var.rName

{{- template "tags" . }}
}
