resource "aws_kinesis_stream_consumer" "test" {
{{- template "region" }}
  name       = var.rName
  stream_arn = aws_kinesis_stream.test.arn

{{- template "tags" . }}
}

resource "aws_kinesis_stream" "test" {
{{- template "region" }}
  name        = var.rName
  shard_count = 2
}