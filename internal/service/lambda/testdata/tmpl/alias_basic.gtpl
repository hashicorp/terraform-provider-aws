resource "aws_lambda_alias" "test" {
{{- template "region" }}
  name             = var.rName
  function_name    = aws_lambda_function.test.function_name
  function_version = aws_lambda_function.test.version
}

resource "aws_lambda_function" "test" {
{{- template "region" }}
  filename      = "test-fixtures/lambdatest.zip"
  function_name = var.rName
  role          = aws_iam_role.test.arn
  handler       = "exports.example"
  runtime       = "nodejs22.x"
  publish       = true
}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}
