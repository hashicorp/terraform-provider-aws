resource "aws_kinesis_stream" "test" {
  name        = "amazon-workspaces-web-${var.rName}"
  shard_count = 1
}

resource "aws_workspacesweb_user_access_logging_settings" "test" {
  kinesis_stream_arn = aws_kinesis_stream.test.arn

{{- template "tags" . }}

}