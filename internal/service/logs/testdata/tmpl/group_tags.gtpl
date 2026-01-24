resource "aws_cloudwatch_log_group" "test" {
{{- template "region" }}
  name = var.rName

  retention_in_days = 1

{{- template "tags" . }}
}
