---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_function_url"
description: |-
  Creates a Lambda function URL.
---

# Resource: aws_lambda_function_url

Creates a Lambda function URL. Creates a https url that points to the specified Lambda function alias or the latest version.

A function URL is a dedicated HTTP(S) endpoint for your function. For information about Lambda function url, see [CreateFunctionUrlConfig][1] in the API docs.

## Example Usage

```terraform
resource "aws_lambda_function_url" "test_latest" {
  function_name    = aws_lambda_function.test.function_name
  authorization_type = "NONE"
}

resource "aws_lambda_function_url" "test_live" {
  function_name    = aws_lambda_function.test.function_name
  qualifier        = "my_alias"
  authorization_type = "AWS_IAM"
  cors {
    allow_credentials = true
    allow_origins     = ["*"]
    allow_methods     = ["*"]
    allow_headers     = ["date", "keep-alive"]
    expose_headers    = ["keep-alive", "date"]
    max_age           = 86400
  }
}
```

## Argument Reference

* `function_name` - (Required) Lambda Function name or ARN.
* `qualifier` - (Optional) The Lambda alias name or '$LATEST'.
* `authorization_type` - (Required) The authorization type for your function URL. Valid values are `["AWS_IAM"]` and `["NONE"]`.
* `cors` - (Optional) Configure cross-origin resource sharing.

### cors

* `allow_credentials` - (Optional) allow cookies or other credentials in requests to your function URL.
* `allow_origins` - (Optional) The origins that can access your function URL. You can list any number of specific origins. Alternatively, you can grant access to all origins with the wildcard character (*). For example: https://www.example.com, https://*, or *.
* `allow_methods` - (Optional) The HTTP methods that are allowed when calling your function URL. For example: GET, POST, DELETE, *.
* `allow_headers` - (Optional) The HTTP headers that origins can include in requests to your function URL. For example: Date, Keep-Alive, X-Custom-Header.
* `expose_headers` - (Optional) The HTTP headers in your function response that you want to expose to origins that call your function URL. For example: Date, Keep-Alive, X-Custom-Header.
* `max_age` - (Optional) The maximum amount of time (in seconds) that browsers can cache results of a preflight request. You cannot exceed 86400 seconds max age.


[1]: http://docs.aws.amazon.com/lambda/latest/dg/API_CreateFunctionUrlConfig.html

## Import

Lambda Function Aliases can be imported using the `function_name/alias`, e.g.,

```
$ terraform import aws_lambda_function_url.test_lambda_url my_test_lambda_function/my_alias
```
