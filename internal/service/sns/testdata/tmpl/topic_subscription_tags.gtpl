resource "aws_sns_topic" "test" {
{{- template "region" }}
  name = var.rName
}

resource "aws_sqs_queue" "test" {
{{- template "region" }}
  name = var.rName

  sqs_managed_sse_enabled = true
}

resource "aws_sns_topic_subscription" "test" {
{{- template "region" }}
  topic_arn = aws_sns_topic.test.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.test.arn
{{- template "tags" }}
}
