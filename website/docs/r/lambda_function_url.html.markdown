---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_function_url"
description: |-
  Provides a Lambda function URL resource.
---

# Resource: aws_lambda_function_url

Provides a Lambda function URL resource. A function URL is a dedicated HTTP(S) endpoint for a Lambda function.

See the [AWS Lambda documentation](https://docs.aws.amazon.com/lambda/latest/dg/lambda-urls.html) for more information.

## Example Usage

```terraform
resource "aws_lambda_function_url" "test_latest" {
  function_name      = aws_lambda_function.test.function_name
  authorization_type = "NONE"
}

resource "aws_lambda_function_url" "test_live" {
  function_name      = aws_lambda_function.test.function_name
  qualifier          = "my_alias"
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

* `authorization_type` - (Required) The type of authentication that the function URL uses. Set to `"AWS_IAM"` to restrict access to authenticated IAM users only. Set to `"NONE"` to bypass IAM authentication and create a public endpoint. See the [AWS documentation](https://docs.aws.amazon.com/lambda/latest/dg/urls-auth.html) for more details.
* `cors` - (Optional) The [cross-origin resource sharing (CORS)](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS) settings for the function URL. Documented below.
* `function_name` - (Required) The name (or ARN) of the Lambda function.
* `invoke_mode` - (Optional) Determines how the Lambda function responds to an invocation. Valid values are `BUFFERED` (default) and `RESPONSE_STREAM`. See more in [Configuring a Lambda function to stream responses](https://docs.aws.amazon.com/lambda/latest/dg/configuration-response-streaming.html).
* `qualifier` - (Optional) The alias name or `"$LATEST"`.

### cors

This configuration block supports the following attributes:

* `allow_credentials` - (Optional) Whether to allow cookies or other credentials in requests to the function URL. The default is `false`.
* `allow_headers` - (Optional) The HTTP headers that origins can include in requests to the function URL. For example: `["date", "keep-alive", "x-custom-header"]`.
* `allow_methods` - (Optional) The HTTP methods that are allowed when calling the function URL. For example: `["GET", "POST", "DELETE"]`, or the wildcard character (`["*"]`).
* `allow_origins` - (Optional) The origins that can access the function URL. You can list any number of specific origins (or the wildcard character (`"*"`)), separated by a comma. For example: `["https://www.example.com", "http://localhost:60905"]`.
* `expose_headers` - (Optional) The HTTP headers in your function response that you want to expose to origins that call the function URL.
* `max_age` - (Optional) The maximum amount of time, in seconds, that web browsers can cache results of a preflight request. By default, this is set to `0`, which means that the browser doesn't cache results. The maximum value is `86400`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `function_arn` - The Amazon Resource Name (ARN) of the function.
* `function_url` - The HTTP URL endpoint for the function in the format `https://<url_id>.lambda-url.<region>.on.aws/`.
* `url_id` - A generated ID for the endpoint.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Lambda function URLs using the `function_name` or `function_name/qualifier`. For example:

```terraform
import {
  to = aws_lambda_function_url.test_lambda_url
  id = "my_test_lambda_function"
}
```

Using `terraform import`, import Lambda function URLs using the `function_name` or `function_name/qualifier`. For example:

```console
% terraform import aws_lambda_function_url.test_lambda_url my_test_lambda_function
```
