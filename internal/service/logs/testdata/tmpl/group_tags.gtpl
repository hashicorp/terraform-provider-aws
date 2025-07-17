resource "aws_cloudwatch_log_group" "test" {
  name = var.rName

  retention_in_days = 1

{{- template "tags" . }}
}
