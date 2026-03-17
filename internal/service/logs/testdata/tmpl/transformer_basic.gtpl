resource "aws_cloudwatch_log_transformer" "test" {
{{- template "region" }}
  log_group_arn = aws_cloudwatch_log_group.test.arn

  transformer_config {
    parse_json {}
  }
}

resource "aws_cloudwatch_log_group" "test" {
{{- template "region" }}
  name = var.rName
}