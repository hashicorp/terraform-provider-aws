resource "aws_sqs_queue" "test" {
{{- template "region" }}
  name = var.rName

{{- template "tags" . }}
}
