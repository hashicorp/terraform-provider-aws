---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_invocation"
description: |-
  Invoke AWS Lambda Function
---

# Resource: aws_lambda_invocation

Use this resource to invoke a lambda function. The lambda function is invoked with the [RequestResponse](https://docs.aws.amazon.com/lambda/latest/dg/API_Invoke.html#API_Invoke_RequestSyntax) invocation type.

~> **NOTE:** By default this resource _only_ invokes the function when the arguments call for a create or replace. In other words, after an initial invocation on _apply_, if the arguments do not change, a subsequent _apply_ does not invoke the function again. To dynamically invoke the function, see the `triggers` example below. To always invoke a function on each _apply_, see the [`aws_lambda_invocation`](/docs/providers/aws/d/lambda_invocation.html) data source. To invoke the lambda function when the terraform resource is updated and deleted, see the [CRUD Lifecycle Scope](#crud-lifecycle-scope) example below.

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

### CRUD Lifecycle Scope

```terraform
resource "aws_lambda_invocation" "example" {
  function_name = aws_lambda_function.lambda_function_test.function_name

  input = jsonencode({
    key1 = "value1"
    key2 = "value2"
  })

  lifecycle_scope = "CRUD"
}
```

~> **NOTE:** `lifecycle_scope = "CRUD"` will inject a key `tf` in the input event to pass lifecycle information! This allows the lambda function to handle different lifecycle transitions uniquely.  If you need to use a key `tf` in your own input JSON, the default key name can be overridden with the `terraform_key` argument.

The key `tf` gets added with subkeys:

* `action` - Action Terraform performs on the resource. Values are `create`, `update`, or `delete`.
* `prev_input` - Input JSON payload from the previous invocation. This can be used to handle update and delete events.

When the resource from the example above is created, the Lambda will get following JSON payload:

```json
{
  "key1": "value1",
  "key2": "value2",
  "tf": {
    "action": "create",
    "prev_input": null
  }
}
```

If the input value of `key1` changes to "valueB", then the lambda will be invoked again with the following JSON payload:

```json
{
  "key1": "valueB",
  "key2": "value2",
  "tf": {
    "action": "update",
    "prev_input": {
      "key1": "value1",
      "key2": "value2"
    }
  }
}
```

When the invocation resource is removed, the final invocation will have the following JSON payload:

```json
{
  "key1": "valueB",
  "key2": "value2",
  "tf": {
    "action": "delete",
    "prev_input": {
      "key1": "valueB",
      "key2": "value2"
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `function_name` - (Required) Name of the lambda function.
* `input` - (Required) JSON payload to the lambda function.

The following arguments are optional:

* `lifecycle_scope` - (Optional) Lifecycle scope of the resource to manage. Valid values are `CREATE_ONLY` and `CRUD`. Defaults to `CREATE_ONLY`. `CREATE_ONLY` will invoke the function only on creation or replacement. `CRUD` will invoke the function on each lifecycle event, and augment the input JSON payload with additional lifecycle information.
* `qualifier` - (Optional) Qualifier (i.e., version) of the lambda function. Defaults to `$LATEST`.
* `terraform_key` - (Optional) The JSON key used to store lifecycle information in the input JSON payload. Defaults to `tf`. This additional key is only included when `lifecycle_scope` is set to `CRUD`.
* `triggers` - (Optional) Map of arbitrary keys and values that, when changed, will trigger a re-invocation. To force a re-invocation without changing these keys/values, use the [`terraform taint` command](https://www.terraform.io/docs/commands/taint.html).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `result` - String result of the lambda function invocation.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Lambda Invocation using the `function_name_qualiefier_result_hash`. For example:

```terraform
import {
  to = aws_lambda_invocation.test_lambda
  id = "my_test_lambda_function_$LATEST_b326b5062b2f0e69046810717534cb09"
}
```

Using `terraform import`, import Lambda Functions using the `function_name`. For example:

```console
% terraform import my_test_lambda_function_$LATEST_b326b5062b2f0e69046810717534cb09
```

Because it is not possible to retrieve previous invocations, during the next apply `terraform` will update the resource calling again the function.
To compute the `result_hash`, it is necessary to hash it with the standard `md5` hash function.
