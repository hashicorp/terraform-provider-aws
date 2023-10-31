---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_alias"
description: |-
  Provides a Lambda Alias data source.
---

# Data Source: aws_lambda_alias

Provides information about a Lambda Alias.

## Example Usage

```terraform
data "aws_lambda_alias" "production" {
  function_name = "my-lambda-func"
  name          = "production"
}
```

## Argument Reference

This data source supports the following arguments:

* `function_name` - (Required) Name of the aliased Lambda function.
* `name` - (Required) Name of the Lambda alias.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN identifying the Lambda function alias.
* `description` - Description of alias.
* `function_version` - Lambda function version which the alias uses.
* `invoke_arn` - ARN to be used for invoking Lambda Function from API Gateway - to be used in aws_api_gateway_integration's `uri`.
