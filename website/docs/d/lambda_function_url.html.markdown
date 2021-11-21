---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_function_url"
description: |-
  Provides a Lambda function url data source.
---

# aws_lambda_function_url

Provides information about a Lambda function url.

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

The following arguments are supported:

* `function_name` - (Required) Name of the lambda function.
* `qualifier` - (Optional) Alias name or latest version of the lambda function E.g., `$LATEST`, or `my-alias`

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `authorization_type` - The authorization type for your function URL.
* `cors` - Cross-Origin Resource Sharing configuration.
* `function_arn` - Lambda Function full ARN which this URL points to.
* `function_url` - The HTTPS endpoint of Lambda URL.
* `creation_time` - The creation time.
* `last_modified_time` - The last modified time.