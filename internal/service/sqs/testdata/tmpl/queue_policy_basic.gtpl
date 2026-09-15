resource "aws_sqs_queue_policy" "test" {
{{- template "region" }}
  queue_url = aws_sqs_queue.test.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": "*",
    "Action": "sqs:*",
    "Resource": "${aws_sqs_queue.test.arn}",
    "Condition": {
      "ArnEquals": {
        "aws:SourceArn": "${aws_sqs_queue.test.arn}"
      }
    }
  }]
}
POLICY
}

resource "aws_sqs_queue" "test" {
{{- template "region" }}
  name = var.rName
}
