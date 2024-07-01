resource "aws_sqs_queue" "test" {
  name = var.rName

{{- template "tags" . }}
}
