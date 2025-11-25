# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_secretsmanager_secret_rotation" "test" {
  secret_id           = aws_secretsmanager_secret.test.id
  rotation_lambda_arn = aws_lambda_function.test.arn

  rotation_rules {
    automatically_after_days = 7
  }

  depends_on = [aws_lambda_permission.test]
}

data "aws_partition" "current" {
}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "lambda.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy" "test" {
  name = var.rName
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = [
        "secretsmanager:DescribeSecret",
        "secretsmanager:GetSecretValue",
        "secretsmanager:PutSecretValue",
        "secretsmanager:UpdateSecretVersionStage",
      ]
      Effect   = "Allow"
      Resource = aws_secretsmanager_secret.test.arn
    }]
  })
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
  role       = aws_iam_role.test.name
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = var.rName
  handler       = "exports.example"
  role          = aws_iam_role.test.arn
  runtime       = "nodejs20.x"
}

resource "aws_lambda_permission" "test" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.function_name
  principal     = "secretsmanager.amazonaws.com"
  statement_id  = "AllowExecutionFromSecretsManager1"
}

resource "aws_secretsmanager_secret" "test" {
  name = var.rName
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = "test-string"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
