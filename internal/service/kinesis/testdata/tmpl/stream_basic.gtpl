resource "aws_kinesis_stream" "test" {
{{- template "region" }}
  name        = var.rName
  shard_count = 2

{{- template "tags" . }}
}
