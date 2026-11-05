resource "aws_iam_role" "test" {
  name = var.rName
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Principal = { Service = "lambda.amazonaws.com" }
      Effect    = "Allow"
    }]
  })
}

resource "aws_lambda_function" "test" {
{{- template "region" }}
  filename         = "test-fixtures/lambdatest.zip"
  source_code_hash = filebase64sha256("test-fixtures/lambdatest.zip")
  function_name    = var.rName
  role             = aws_iam_role.test.arn
  handler          = "exports.example"
  runtime          = "nodejs24.x"
}

resource "aws_iot_authorizer" "test" {
{{- template "region" }}
  name                    = var.rName
  authorizer_function_arn = aws_lambda_function.test.arn
  token_key_name          = "Token-Header-1"
  token_signing_public_keys = {
    Key1 = file("test-fixtures/iot-authorizer-signing-key.pem")
  }
{{- template "tags" . }}
}
