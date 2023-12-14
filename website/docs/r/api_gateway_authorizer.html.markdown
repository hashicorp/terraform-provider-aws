---
subcategory: "API Gateway"
layout: "aws"
page_title: "AWS: aws_api_gateway_authorizer"
description: |-
  Provides an API Gateway Authorizer.
---

# Resource: aws_api_gateway_authorizer

Provides an API Gateway Authorizer.

## Example Usage

```terraform
resource "aws_api_gateway_authorizer" "demo" {
  name                   = "demo"
  rest_api_id            = aws_api_gateway_rest_api.demo.id
  authorizer_uri         = aws_lambda_function.authorizer.invoke_arn
  authorizer_credentials = aws_iam_role.invocation_role.arn
}

resource "aws_api_gateway_rest_api" "demo" {
  name = "auth-demo"
}

data "aws_iam_policy_document" "invocation_assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["apigateway.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "invocation_role" {
  name               = "api_gateway_auth_invocation"
  path               = "/"
  assume_role_policy = data.aws_iam_policy_document.invocation_assume_role.json
}

data "aws_iam_policy_document" "invocation_policy" {
  statement {
    effect    = "Allow"
    actions   = ["lambda:InvokeFunction"]
    resources = [aws_lambda_function.authorizer.arn]
  }
}

resource "aws_iam_role_policy" "invocation_policy" {
  name   = "default"
  role   = aws_iam_role.invocation_role.id
  policy = data.aws_iam_policy_document.invocation_policy.json
}

data "aws_iam_policy_document" "lambda_assume_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "lambda" {
  name               = "demo-lambda"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json
}

resource "aws_lambda_function" "authorizer" {
  filename      = "lambda-function.zip"
  function_name = "api_gateway_authorizer"
  role          = aws_iam_role.lambda.arn
  handler       = "exports.example"

  source_code_hash = filebase64sha256("lambda-function.zip")
}
```

## Argument Reference

This resource supports the following arguments:

* `authorizer_uri` - (Optional, required for type `TOKEN`/`REQUEST`) Authorizer's Uniform Resource Identifier (URI). This must be a well-formed Lambda function URI in the form of `arn:aws:apigateway:{region}:lambda:path/{service_api}`,
 e.g., `arn:aws:apigateway:us-west-2:lambda:path/2015-03-31/functions/arn:aws:lambda:us-west-2:012345678912:function:my-function/invocations`
* `name` - (Required) Name of the authorizer
* `rest_api_id` - (Required) ID of the associated REST API
* `identity_source` - (Optional) Source of the identity in an incoming request. Defaults to `method.request.header.Authorization`. For `REQUEST` type, this may be a comma-separated list of values, including headers, query string parameters and stage variables - e.g., `"method.request.header.SomeHeaderName,method.request.querystring.SomeQueryStringName,stageVariables.SomeStageVariableName"`
* `type` - (Optional) Type of the authorizer. Possible values are `TOKEN` for a Lambda function using a single authorization token submitted in a custom header, `REQUEST` for a Lambda function using incoming request parameters, or `COGNITO_USER_POOLS` for using an Amazon Cognito user pool. Defaults to `TOKEN`.
* `authorizer_credentials` - (Optional) Credentials required for the authorizer. To specify an IAM Role for API Gateway to assume, use the IAM Role ARN.
* `authorizer_result_ttl_in_seconds` - (Optional) TTL of cached authorizer results in seconds. Defaults to `300`.
* `identity_validation_expression` - (Optional) Validation expression for the incoming identity. For `TOKEN` type, this value should be a regular expression. The incoming token from the client is matched against this expression, and will proceed if the token matches. If the token doesn't match, the client receives a 401 Unauthorized response.
* `provider_arns` - (Optional, required for type `COGNITO_USER_POOLS`) List of the Amazon Cognito user pool ARNs. Each element is of this format: `arn:aws:cognito-idp:{region}:{account_id}:userpool/{user_pool_id}`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the API Gateway Authorizer
* `id` - Authorizer identifier.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AWS API Gateway Authorizer using the `REST-API-ID/AUTHORIZER-ID`. For example:

```terraform
import {
  to = aws_api_gateway_authorizer.authorizer
  id = "12345abcde/example"
}
```

Using `terraform import`, import AWS API Gateway Authorizer using the `REST-API-ID/AUTHORIZER-ID`. For example:

```console
% terraform import aws_api_gateway_authorizer.authorizer 12345abcde/example
```
