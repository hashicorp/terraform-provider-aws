resource "aws_sqs_queue_redrive_policy" "test" {
{{- template "region" }}
  queue_url = aws_sqs_queue.test.id
  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.test_ddl.arn
    maxReceiveCount     = 4
  })
}

resource "aws_sqs_queue" "test" {
{{- template "region" }}
  name = var.rName
}

resource "aws_sqs_queue" "test_ddl" {
{{- template "region" }}
  name = "${var.rName}_ddl"
  redrive_allow_policy = jsonencode({
    redrivePermission = "byQueue",
    sourceQueueArns   = [aws_sqs_queue.test.arn]
  })
}
