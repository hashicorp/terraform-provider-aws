---
subcategory: "Connect"
layout: "aws"
page_title: "AWS: aws_connect_lambda_function_association"
description: |-
  Provides details about a specific Connect Lambda Function Association.
---

# Resource: aws_connect_lambda_function_association

Provides an Amazon Connect Lambda Function Association. For more information see
[Amazon Connect: Getting Started](https://docs.aws.amazon.com/connect/latest/adminguide/amazon-connect-get-started.html)

[Invoke AWS Lambda functions](https://docs.aws.amazon.com/connect/latest/adminguide/connect-lambda-functions.html)

## Example Usage

```terraform
resource "aws_connect_lambda_function_association" "test" {
  function_arn = aws_lambda_function.test.arn
  instance_id  = aws_connect_instance.test.id
}
```

## Argument Reference

The following arguments are supported:

* `function_arn` - (Required) Amazon Resource Name (ARN) of the Lambda Function, omitting any version or alias qualifier.
* `instance_id` - (Required) The identifier of the Amazon Connect instance. You can find the instanceId in the ARN of the instance.

### Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 1 mins) Used when creating the association.
* `delete` - (Defaults to 1 mins) Used when creating the association.

## Attributes Reference

No additional attributes are exported.
