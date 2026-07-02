resource "aws_config_organization_custom_rule" "test" {
{{- template "region" }}
  depends_on = [aws_config_configuration_recorder.test, aws_lambda_permission.test, aws_organizations_organization.test]

  lambda_function_arn = aws_lambda_function.test.arn
  name                = var.rName
  trigger_types       = ["ConfigurationItemChangeNotification"]
}

data "aws_partition" "current" {}

resource "aws_config_configuration_recorder" "test" {
{{- template "region" }}
  depends_on = [aws_iam_role_policy_attachment.config]

  name     = var.rName
  role_arn = aws_iam_role.config.arn
}

resource "aws_iam_role" "config" {
  name = "${var.rName}-config"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "config.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "config" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWS_ConfigRole"
  role       = aws_iam_role.config.name
}

resource "aws_iam_role" "lambda" {
  name = "${var.rName}-lambda"

  assume_role_policy = <<POLICY
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
POLICY
}

resource "aws_iam_role_policy_attachment" "lambda" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSConfigRulesExecutionRole"
  role       = aws_iam_role.lambda.name
}

resource "aws_lambda_function" "test" {
{{- template "region" }}
  filename      = "lambdatest.zip"
  function_name = var.rName
  role          = aws_iam_role.lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs24.x"
}

resource "aws_lambda_permission" "test" {
{{- template "region" }}
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.arn
  principal     = "config.amazonaws.com"
  statement_id  = "AllowExecutionFromConfig"
}

resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["config-multiaccountsetup.amazonaws.com"]
  feature_set                   = "ALL"
}