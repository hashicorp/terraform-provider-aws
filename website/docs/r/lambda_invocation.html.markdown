---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_invocation"
description: |-
  Invoke AWS Lambda Function
---

# Resource: aws_lambda_invocation

Use this resource to invoke a lambda function. The lambda function is invoked with the [RequestResponse](https://docs.aws.amazon.com/lambda/latest/dg/API_Invoke.html#API_Invoke_RequestSyntax) invocation type.

~> **NOTE:** By default this resource _only_ invokes the function when the arguments call for a create or replace. In other words, after an initial invocation on _apply_, if the arguments do not change, a subsequent _apply_ does not invoke the function again. To dynamically invoke the function, see the `triggers` example below. To always invoke a function on each _apply_, see the [`aws_lambda_invocation`](/docs/providers/aws/d/lambda_invocation.html) data source. To invoke the lambda function when the terraform resource is updated/deleted see the `CRUD` example below.

## Example Usage

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

### CRUD lifecycle_scope: Process all the lifecycle events on the terraform resource

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

~> **NOTE:** `lifecycle_scope = "CRUD"` will inject a key `tf` in the input event to pass lifecycle information! This allows
to implement logic in your lambda function to handle different lifecycle transitions uniquely.  If you need to use a key `tf` in your own input JSON then see the `terraform_key` argument.

The key `tf` gets added with subkeys:
 * `action` which gets a value corresponding to the action terraform performs on the resource [`create`, `delete`, `update`]
 * `prev_input` which gets a value of the previous invocation. This can be used to handle deletes and updates.

When the resource from the example above gets added the Lambda will get following JSON payload:
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

So if for the above example the input would change the value of `key1` to "valueB" then the lambda will be invoked once more with the following JSON body:

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

If finally the lambda_invocation resource will be removed then a final invocation happens with JSON body:

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

* `lifecycle_scope` - (Optional) Lifecycle scope of the resource to manage. Supported values:
  * `"CREATE_ONLY"` (Default): Trigger only on create or replace of the terraform resources.
  * `"CRUD"`: Manage the full lifecycle and augment JSON payload of the lambda function with lifecycle information.
* `qualifier` - (Optional) Qualifier (i.e., version) of the lambda function. Defaults to `$LATEST`.
* `terraform_key` - (Optional) The JSON key used to store lifecycle information in the JSON payload for the lambda function. Default "tf".
* `triggers` - (Optional) Map of arbitrary keys and values that, when changed, will trigger a re-invocation. To force a re-invocation without changing these keys/values, use the [`terraform taint` command](https://www.terraform.io/docs/commands/taint.html).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `result` - String result of the lambda function invocation.
