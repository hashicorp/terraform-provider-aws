---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_function_url"
description: Manages a Lambda function URL.
---

# Resource: aws_lambda_function_url

Manages a Lambda function URL. Creates a dedicated HTTP(S) endpoint for a Lambda function to enable direct invocation via HTTP requests.

## Example Usage

### Basic Function URL with No Authentication

```terraform
resource "aws_lambda_function_url" "example" {
  function_name      = aws_lambda_function.example.function_name
  authorization_type = "NONE"
}
```

### Function URL with IAM Authentication and CORS Configuration

```terraform
resource "aws_lambda_function_url" "example" {
  function_name      = aws_lambda_function.example.function_name
  qualifier          = "my_alias"
  authorization_type = "AWS_IAM"
  invoke_mode        = "RESPONSE_STREAM"

  cors {
    allow_credentials = true
    allow_origins     = ["https://example.com"]
    allow_methods     = ["GET", "POST"]
    allow_headers     = ["date", "keep-alive"]
    expose_headers    = ["keep-alive", "date"]
    max_age           = 86400
  }
}
```

## Argument Reference

The following arguments are required:

* `authorization_type` - (Required) Type of authentication that the function URL uses. Valid values are `AWS_IAM` and `NONE`.
* `function_name` - (Required) Name or ARN of the Lambda function.

The following arguments are optional:

* `cors` - (Optional) Cross-origin resource sharing (CORS) settings for the function URL. [See below](#cors).
* `invoke_mode` - (Optional) How the Lambda function responds to an invocation. Valid values are `BUFFERED` (default) and `RESPONSE_STREAM`.
* `qualifier` - (Optional) Alias name or `$LATEST`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### CORS

* `allow_credentials` - (Optional) Whether to allow cookies or other credentials in requests to the function URL.
* `allow_headers` - (Optional) HTTP headers that origins can include in requests to the function URL.
* `allow_methods` - (Optional) HTTP methods that are allowed when calling the function URL.
* `allow_origins` - (Optional) Origins that can access the function URL.
* `expose_headers` - (Optional) HTTP headers in your function response that you want to expose to origins that call the function URL.
* `max_age` - (Optional) Maximum amount of time, in seconds, that web browsers can cache results of a preflight request. Maximum value is `86400`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `function_arn` - ARN of the Lambda function.
* `function_url` - HTTP URL endpoint for the function in the format `https://<url_id>.lambda-url.<region>.on.aws/`.
* `url_id` - Generated ID for the endpoint.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Lambda function URLs using the `function_name` or `function_name/qualifier`. For example:

```terraform
import {
  to = aws_lambda_function_url.example
  id = "example"
}
```

Using `terraform import`, import Lambda function URLs using the `function_name` or `function_name/qualifier`. For example:

```console
% terraform import aws_lambda_function_url.example example
```
