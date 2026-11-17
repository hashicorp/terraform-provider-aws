resource "aws_lambdaweb_resource_policy" "test" {
{{- template "region" }}
  resource_arn = aws_lambdaweb_function.test.arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid    = "AllowAccount"
      Effect = "Allow"
      Principal = {
        AWS = data.aws_caller_identity.current.account_id
      }
      Action   = "lambda:InvokeWebFunctionEndpoint"
      Resource = "${aws_lambdaweb_function.test.arn}/endpoint/default"
    }]
  })
}

resource "aws_lambdaweb_function" "test" {
{{- template "region" }}
  function_name = var.rName
}

data "aws_caller_identity" "current" {}
