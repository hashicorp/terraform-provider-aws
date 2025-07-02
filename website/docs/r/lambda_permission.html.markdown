---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_permission"
description: |-
  Manages an AWS Lambda permission.
---

# Resource: aws_lambda_permission

Manages an AWS Lambda permission. Use this resource to grant external sources (e.g., EventBridge Rules, SNS, or S3) permission to invoke Lambda functions.

## Example Usage

### Basic Usage with EventBridge

```terraform
resource "aws_lambda_permission" "allow_cloudwatch" {
  statement_id  = "AllowExecutionFromCloudWatch"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test_lambda.function_name
  principal     = "events.amazonaws.com"
  source_arn    = "arn:aws:events:eu-west-1:111122223333:rule/RunDaily"
  qualifier     = aws_lambda_alias.test_alias.name
}

resource "aws_lambda_alias" "test_alias" {
  name             = "testalias"
  description      = "a sample description"
  function_name    = aws_lambda_function.test_lambda.function_name
  function_version = "$LATEST"
}

resource "aws_lambda_function" "test_lambda" {
  filename      = "lambdatest.zip"
  function_name = "lambda_function_name"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.handler"
  runtime       = "nodejs20.x"
}

resource "aws_iam_role" "iam_for_lambda" {
  name = "iam_for_lambda"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}
```

### SNS Integration

```terraform
resource "aws_lambda_permission" "with_sns" {
  statement_id  = "AllowExecutionFromSNS"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.func.function_name
  principal     = "sns.amazonaws.com"
  source_arn    = aws_sns_topic.default.arn
}

resource "aws_sns_topic" "default" {
  name = "call-lambda-maybe"
}

resource "aws_sns_topic_subscription" "lambda" {
  topic_arn = aws_sns_topic.default.arn
  protocol  = "lambda"
  endpoint  = aws_lambda_function.func.arn
}

resource "aws_lambda_function" "func" {
  filename      = "lambdatest.zip"
  function_name = "lambda_called_from_sns"
  role          = aws_iam_role.default.arn
  handler       = "exports.handler"
  runtime       = "python3.12"
}

resource "aws_iam_role" "default" {
  name = "iam_for_lambda_with_sns"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}
```

### API Gateway REST API Integration

```terraform
resource "aws_api_gateway_rest_api" "MyDemoAPI" {
  name        = "MyDemoAPI"
  description = "This is my API for demonstration purposes"
}

resource "aws_lambda_permission" "lambda_permission" {
  statement_id  = "AllowMyDemoAPIInvoke"
  action        = "lambda:InvokeFunction"
  function_name = "MyDemoFunction"
  principal     = "apigateway.amazonaws.com"

  # The /* part allows invocation from any stage, method and resource path
  # within API Gateway.
  source_arn = "${aws_api_gateway_rest_api.MyDemoAPI.execution_arn}/*"
}
```

### CloudWatch Log Group Integration

```terraform
resource "aws_lambda_permission" "logging" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.logging.function_name
  principal     = "logs.eu-west-1.amazonaws.com"
  source_arn    = "${aws_cloudwatch_log_group.default.arn}:*"
}

resource "aws_cloudwatch_log_group" "default" {
  name = "/default"
}

resource "aws_cloudwatch_log_subscription_filter" "logging" {
  depends_on      = [aws_lambda_permission.logging]
  destination_arn = aws_lambda_function.logging.arn
  filter_pattern  = ""
  log_group_name  = aws_cloudwatch_log_group.default.name
  name            = "logging_default"
}

resource "aws_lambda_function" "logging" {
  filename      = "lamba_logging.zip"
  function_name = "lambda_called_from_cloudwatch_logs"
  handler       = "exports.handler"
  role          = aws_iam_role.default.arn
  runtime       = "python3.12"
}

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

resource "aws_iam_role" "default" {
  name               = "iam_for_lambda_called_from_cloudwatch_logs"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}
```

### Cross-Account Function URL Access

```terraform
resource "aws_lambda_function_url" "url" {
  function_name      = aws_lambda_function.example.function_name
  authorization_type = "AWS_IAM"
}

resource "aws_lambda_permission" "url" {
  action        = "lambda:InvokeFunctionUrl"
  function_name = aws_lambda_function.example.function_name
  principal     = "arn:aws:iam::444455556666:role/example"

  source_account         = "444455556666"
  function_url_auth_type = "AWS_IAM"
}
```

### Automatic Permission Updates with Function Changes

```terraform
resource "aws_lambda_permission" "logging" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.example.function_name
  principal     = "events.amazonaws.com"
  source_arn    = "arn:aws:events:eu-west-1:111122223333:rule/RunDaily"

  lifecycle {
    replace_triggered_by = [
      aws_lambda_function.example
    ]
  }
}
```

## Argument Reference

The following arguments are required:

* `action` - (Required) Lambda action to allow in this statement (e.g., `lambda:InvokeFunction`)
* `function_name` - (Required) Name of the Lambda function
* `principal` - (Required) AWS service or account that invokes the function (e.g., `s3.amazonaws.com`, `sns.amazonaws.com`, AWS account ID, or AWS IAM principal)

The following arguments are optional:

* `event_source_token` - (Optional) Event Source Token for Alexa Skills
* `function_url_auth_type` - (Optional) Lambda Function URL authentication type. Valid values: `AWS_IAM` or `NONE`. Only valid with `lambda:InvokeFunctionUrl` action
* `principal_org_id` - (Optional) AWS Organizations ID to grant permission to all accounts under this organization
* `qualifier` - (Optional) Lambda function version or alias name
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference)
* `source_account` - (Optional) AWS account ID of the source owner for cross-account access, S3, or SES
* `source_arn` - (Optional) ARN of the source resource granting permission to invoke the Lambda function
* `statement_id` - (Optional) Statement identifier. Generated by Terraform if not provided
* `statement_id_prefix` - (Optional) Statement identifier prefix. Conflicts with `statement_id`

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Lambda permission statements using function_name/statement_id with an optional qualifier. For example:

```terraform
import {
  to = aws_lambda_permission.test_lambda_permission
  id = "my_test_lambda_function/AllowExecutionFromCloudWatch"
}
```

Using `qualifier`:

```terraform
import {
  to = aws_lambda_permission.test_lambda_permission
  id = "my_test_lambda_function:qualifier_name/AllowExecutionFromCloudWatch"
}
```

For backwards compatibility, the following legacy `terraform import` commands are also supported:

```console
% terraform import aws_lambda_permission.example my_test_lambda_function/AllowExecutionFromCloudWatch
% terraform import aws_lambda_permission.test_lambda_permission my_test_lambda_function:qualifier_name/AllowExecutionFromCloudWatch
```
