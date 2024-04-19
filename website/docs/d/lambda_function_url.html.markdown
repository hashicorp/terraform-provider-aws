---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_function_url"
description: |-
  Provides a Lambda function URL data source.
---

# aws_lambda_function_url

Provides information about a Lambda function URL.

## Example Usage

```terraform
variable "function_name" {
  type = string
}

data "aws_lambda_function_url" "existing" {
  function_name = var.function_name
}
```

## Argument Reference

This data source supports the following arguments:

* `function_name` - (Required) The name (or ARN) of the Lambda function.
* `qualifier` - (Optional) Alias name or `"$LATEST"`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `authorization_type` - Type of authentication that the function URL uses.
* `cors` - The [cross-origin resource sharing (CORS)](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS) settings for the function URL. See the [`aws_lambda_function_url` resource](/docs/providers/aws/r/lambda_function_url.html) documentation for more details.
* `creation_time` - When the function URL was created, in [ISO-8601 format](https://www.w3.org/TR/NOTE-datetime).
* `function_arn` - ARN of the function.
* `function_url` - HTTP URL endpoint for the function in the format `https://<url_id>.lambda-url.<region>.on.aws/`.
* `invoke_mode` - Whether the Lambda function responds in `BUFFERED` or `RESPONSE_STREAM` mode.
* `last_modified_time` - When the function URL configuration was last updated, in [ISO-8601 format](https://www.w3.org/TR/NOTE-datetime).
* `url_id` - Generated ID for the endpoint.
