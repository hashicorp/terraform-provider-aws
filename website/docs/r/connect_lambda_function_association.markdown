---
subcategory: "Connect"
layout: "aws"
page_title: "AWS: aws_connect_lambda_function_association"
description: |-
  Provides details about a specific Connect Lambda Function Association.
---

# Resource: aws_connect_lambda_function_association

Provides an Amazon Connect Lambda Function Association. For more information see
[Amazon Connect: Getting Started](https://docs.aws.amazon.com/connect/latest/adminguide/amazon-connect-get-started.html) and [Invoke AWS Lambda functions](https://docs.aws.amazon.com/connect/latest/adminguide/connect-lambda-functions.html).

## Example Usage

```terraform
resource "aws_connect_lambda_function_association" "example" {
  function_arn = aws_lambda_function.example.arn
  instance_id  = aws_connect_instance.example.id
}
```

## Argument Reference

This resource supports the following arguments:

* `function_arn` - (Required) Amazon Resource Name (ARN) of the Lambda Function, omitting any version or alias qualifier.
* `instance_id` - (Required) The identifier of the Amazon Connect instance. You can find the instanceId in the ARN of the instance.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The Amazon Connect instance ID and Lambda Function ARN separated by a comma (`,`).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_connect_lambda_function_association` using the `instance_id` and `function_arn` separated by a comma (`,`). For example:

```terraform
import {
  to = aws_connect_lambda_function_association.example
  id = "aaaaaaaa-bbbb-cccc-dddd-111111111111,arn:aws:lambda:us-west-2:123456789123:function:example"
}
```

Using `terraform import`, import `aws_connect_lambda_function_association` using the `instance_id` and `function_arn` separated by a comma (`,`). For example:

```console
% terraform import aws_connect_lambda_function_association.example aaaaaaaa-bbbb-cccc-dddd-111111111111,arn:aws:lambda:us-west-2:123456789123:function:example
```
