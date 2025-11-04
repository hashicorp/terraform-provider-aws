---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_function_url"
description: |-
  Provides details about an AWS Lambda Function URL.
---

# Data Source: aws_lambda_function_url

Provides details about an AWS Lambda Function URL. Use this data source to retrieve information about an existing function URL configuration.

## Example Usage

### Basic Usage

```terraform
data "aws_lambda_function_url" "example" {
  function_name = "my_lambda_function"
}

output "function_url" {
  value = data.aws_lambda_function_url.example.function_url
}
```

### With Qualifier

```terraform
data "aws_lambda_function_url" "example" {
  function_name = aws_lambda_function.example.function_name
  qualifier     = "production"
}

# Use the URL in other resources
resource "aws_route53_record" "lambda_alias" {
  zone_id = aws_route53_zone.example.zone_id
  name    = "api.example.com"
  type    = "CNAME"
  ttl     = 300
  records = [replace(data.aws_lambda_function_url.example.function_url, "https://", "")]
}
```

### Retrieve CORS Configuration

```terraform
data "aws_lambda_function_url" "example" {
  function_name = "api_function"
}

locals {
  cors_config     = length(data.aws_lambda_function_url.example.cors) > 0 ? data.aws_lambda_function_url.example.cors[0] : null
  allowed_origins = local.cors_config != null ? local.cors_config.allow_origins : []
}

output "cors_allowed_origins" {
  value = local.allowed_origins
}
```

## Argument Reference

The following arguments are required:

* `function_name` - (Required) Name or ARN of the Lambda function.

The following arguments are optional:

* `qualifier` - (Optional) Alias name or `$LATEST`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `authorization_type` - Type of authentication that the function URL uses.
* `cors` - Cross-origin resource sharing (CORS) settings for the function URL. [See below](#cors-attribute-reference).
* `creation_time` - When the function URL was created, in [ISO-8601 format](https://www.w3.org/TR/NOTE-datetime).
* `function_arn` - ARN of the function.
* `function_url` - HTTP URL endpoint for the function in the format `https://<url_id>.lambda-url.<region>.on.aws/`.
* `invoke_mode` - Whether the Lambda function responds in `BUFFERED` or `RESPONSE_STREAM` mode.
* `last_modified_time` - When the function URL configuration was last updated, in [ISO-8601 format](https://www.w3.org/TR/NOTE-datetime).
* `url_id` - Generated ID for the endpoint.

### cors Attribute Reference

* `allow_credentials` - Whether credentials are included in the CORS request.
* `allow_headers` - List of headers that are specified in the Access-Control-Request-Headers header.
* `allow_methods` - List of HTTP methods that are allowed when calling the function URL.
* `allow_origins` - List of origins that are allowed to make requests to the function URL.
* `expose_headers` - List of headers in the response that you want to expose to the origin that called the function URL.
* `max_age` - Maximum amount of time, in seconds, that web browsers can cache results of a preflight request.
