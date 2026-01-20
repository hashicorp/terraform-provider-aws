data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "test" {
  name               = var.rName
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

data "aws_partition" "current" {
}

resource "aws_lambda_function" "test" {
{{- template "region" }}
  filename         = "test-fixtures/lambdatest.zip"
  function_name    = var.rName
  role             = aws_iam_role.test.arn
  handler          = "exports.example"
  runtime          = "nodejs20.x"
  source_code_hash = filebase64sha256("test-fixtures/lambdatest.zip")
  publish          = "true"
}

resource "aws_lambda_alias" "test" {
{{- template "region" }}
  name             = var.rName
  description      = "a sample description"
  function_name    = aws_lambda_function.test.arn
  function_version = "1"
{{- template "tags" }}
}
