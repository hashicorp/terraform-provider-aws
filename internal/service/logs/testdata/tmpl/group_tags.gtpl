resource "aws_cloudwatch_log_group" "test" {
  name = var.rName

{{- template "tags" . }}
}
