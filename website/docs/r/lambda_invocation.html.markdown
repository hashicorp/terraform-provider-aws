---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_invocation"
description: |-
  Invoke AWS Lambda Function
---

# Resource: aws_lambda_invocation

Use this resource to invoke a lambda function. The lambda function is invoked with the [RequestResponse](https://docs.aws.amazon.com/lambda/latest/dg/API_Invoke.html#API_Invoke_RequestSyntax) invocation type.

~> **NOTE:** This resource _only_ invokes the function when the arguments call for a create or update. In other words, after an initial invocation on _apply_, if the arguments do not change, a subsequent _apply_ does not invoke the function again. To dynamically invoke the function, see the `triggers` example below. To always invoke a function on each _apply_, see the [`aws_lambda_invocation`](/docs/providers/aws/d/lambda_invocation.html) data source.

~> **NOTE:** If you get a `KMSAccessDeniedException: Lambda was unable to decrypt the environment variables because KMS access was denied` error when invoking an [`aws_lambda_function`](/docs/providers/aws/r/lambda_function.html) with environment variables, the IAM role associated with the function may have been deleted and recreated _after_ the function was created. You can fix the problem two ways: 1) updating the function's role to another role and then updating it back again to the recreated role, or 2) by using Terraform to `taint` the function and `apply` your configuration again to recreate the function. (When you create a function, Lambda grants permissions on the KMS key to the function's IAM role. If the IAM role is recreated, the grant is no longer valid. Changing the function's role or recreating the function causes Lambda to update the grant.)

## Example Usage

### Basic Example

```terraform
resource "aws_lambda_invocation" "example" {
  function_name = aws_lambda_function.lambda_function_test.function_name

  input = jsonencode({
    key1 = "value1"
    key2 = "value2"
  })
}

output "result_entry" {
  value = jsondecode(aws_lambda_invocation.example.result)["key1"]
}
```

### Dynamic Invocation Example Using Triggers

```terraform
resource "aws_lambda_invocation" "example" {
  function_name = aws_lambda_function.lambda_function_test.function_name

  triggers = {
    redeployment = sha1(jsonencode([
      aws_lambda_function.example.environment
    ]))
  }

  input = jsonencode({
    key1 = "value1"
    key2 = "value2"
  })
}
```

## Argument Reference

The following arguments are required:

* `function_name` - (Required) Name of the lambda function.
* `input` - (Required) JSON payload to the lambda function.

The following arguments are optional:

* `qualifier` - (Optional) Qualifier (i.e., version) of the lambda function. Defaults to `$LATEST`.
* `triggers` - (Optional) Map of arbitrary keys and values that, when changed, will trigger a re-invocation. To force a re-invocation without changing these keys/values, use the [`terraform taint` command](https://www.terraform.io/docs/commands/taint.html).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `result` - String result of the lambda function invocation.
