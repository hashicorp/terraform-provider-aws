resource "aws_lambda_resource_policy" "test" {
{{- template "region" }}
  resource_arn = aws_lambda_function.test.arn
  policy       = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    sid    = "AllowInvoke"
    effect = "Allow"
    actions = [
      "lambda:InvokeFunction",
    ]
    principals {
      type        = "Service"
      identifiers = ["s3.amazonaws.com"]
    }
    resources = [
      aws_lambda_function.test.arn,
    ]
    condition {
      test     = "StringEquals"
      variable = "aws:SourceAccount"
      values   = [data.aws_caller_identity.current.account_id]
    }
  }
}

data "aws_caller_identity" "current" {}

resource "aws_lambda_function" "test" {
{{- template "region" }}
  filename      = "test-fixtures/lambdatest.zip"
  function_name = var.rName
  role          = aws_iam_role.test.arn
  handler       = "exports.example"
  runtime       = "nodejs22.x"
}

data "aws_partition" "current" {}

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

resource "aws_iam_role_policy" "test" {
  name = var.rName
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:${data.aws_partition.current.partition}:logs:*:*:*"
    }
  ]
}
EOF
}
