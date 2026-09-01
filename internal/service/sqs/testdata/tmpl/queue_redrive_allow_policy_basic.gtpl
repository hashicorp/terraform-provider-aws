resource "aws_sqs_queue_redrive_allow_policy" "test" {
{{- template "region" }}
  queue_url = aws_sqs_queue.test.id
  redrive_allow_policy = jsonencode({
    redrivePermission = "byQueue",
    sourceQueueArns   = [aws_sqs_queue.test_src.arn]
  })
}

resource "aws_sqs_queue" "test" {
{{- template "region" }}
  name = var.rName
}

resource "aws_sqs_queue" "test_src" {
{{- template "region" }}
  name = "${var.rName}_src"
  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.test.arn
    maxReceiveCount     = 4
  })
}
