resource "aws_cloudfront_connection_group" "test" {
  name = var.rName

{{- template "tags" . }}
}
