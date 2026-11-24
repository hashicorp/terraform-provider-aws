resource "aws_cloudwatchomni_view" "test" {
{{- template "region" }}
  name       = "view.${var.rName}"
  definition = "SELECT `@timestamp`, `@message` FROM \"logs.default\""
{{- template "tags" . }}
}
