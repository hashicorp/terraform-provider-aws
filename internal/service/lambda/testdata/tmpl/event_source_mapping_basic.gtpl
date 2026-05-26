resource "aws_lambda_event_source_mapping" "test" {
{{- template "region" }}
  event_source_arn = aws_sqs_queue.test.arn
  function_name    = aws_lambda_function.test.arn

{{- template "tags" . }}
}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {
      "Service": "lambda.amazonaws.com"
    },
    "Effect": "Allow"
  }]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = var.rName
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "sqs:*"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_sqs_queue" "test" {
{{- template "region" }}
  name = var.rName
}

resource "aws_lambda_function" "test" {
{{- template "region" }}
  filename      = "test-fixtures/lambdatest.zip"
  function_name = var.rName
  handler       = "exports.example"
  role          = aws_iam_role.test.arn
  runtime       = "nodejs24.x"
}
