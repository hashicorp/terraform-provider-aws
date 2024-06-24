---
subcategory: "Connect"
layout: "aws"
page_title: "AWS: aws_connect_lambda_function_association"
description: |-
  Provides details about a specific Connect Lambda Function Association.
---

# Data Source: aws_connect_lambda_function_association

Provides details about a specific Connect Lambda Function Association.

## Example Usage

```terraform
data "aws_connect_lambda_function_association" "example" {
  function_arn = "arn:aws:lambda:us-west-2:123456789123:function:abcdefg"
  instance_id  = "aaaaaaaa-bbbb-cccc-dddd-111111111111"
}
```

## Argument Reference

This data source supports the following arguments:

* `function_arn` - (Required) ARN of the Lambda Function, omitting any version or alias qualifier.
* `instance_id` - (Required) Identifier of the Amazon Connect instance. You can find the instanceId in the ARN of the instance.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - AWS Region.
