resource "aws_cloudwatch_log_stream" "test" {
{{- template "region" }}
  name           = "${var.rName}-s"
  log_group_name = aws_cloudwatch_log_group.test.id
}

resource "aws_cloudwatch_log_group" "test" {
{{- template "region" }}
  name = "${var.rName}-g"
}
