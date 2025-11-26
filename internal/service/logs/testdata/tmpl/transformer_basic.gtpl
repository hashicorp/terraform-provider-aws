resource "aws_cloudwatch_log_transformer" "test" {
{{- template "region" }}
  log_group_identifier = aws_cloudwatch_log_group.test.name

  transformer_config {
    parse_json {}
  }
}

resource "aws_cloudwatch_log_group" "test" {
{{- template "region" }}
  name = var.rName
}