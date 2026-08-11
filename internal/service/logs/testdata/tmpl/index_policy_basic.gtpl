resource "aws_cloudwatch_log_index_policy" "test" {
{{- template "region" }}
  log_group_name  = aws_cloudwatch_log_group.test.name
  policy_document = "{\"Fields\":[\"eventName\"]}"
}

resource "aws_cloudwatch_log_group" "test" {
{{- template "region" }}
  name = "/aws/testacc/index-policy-${var.rName}"
}
