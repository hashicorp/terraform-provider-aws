resource "aws_cloudwatch_event_permission" "test" {
{{- template "region" }}
  principal    = "111111111111"
  statement_id = var.rName
}
