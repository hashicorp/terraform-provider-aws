---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_runtime_management_config"
description: |-
  Terraform resource for managing an AWS Lambda Runtime Management Config.
---
# Resource: aws_lambda_runtime_management_config

Terraform resource for managing an AWS Lambda Runtime Management Config.

Refer to the [AWS Lambda documentation](https://docs.aws.amazon.com/lambda/latest/dg/lambda-runtimes.html) for supported runtimes.

~> Deletion of this resource returns the runtime update mode to `Auto` (the default behavior).
To leave the configured runtime management options in-place, use a [`removed` block](https://developer.hashicorp.com/terraform/language/resources/syntax#removing-resources) with the destroy lifecycle set to `false`.

## Example Usage

### Basic Usage

```terraform
resource "aws_lambda_runtime_management_config" "example" {
  function_name     = aws_lambda_function.test.function_name
  update_runtime_on = "FunctionUpdate"
}
```

### `Manual` Update

```terraform
resource "aws_lambda_runtime_management_config" "example" {
  function_name     = aws_lambda_function.test.function_name
  update_runtime_on = "Manual"

  # Runtime version ARN's contain a hashed value (not the friendly runtime
  # name). There are currently no API's to retrieve this ARN, but the value
  # can be copied from the "Runtime settings" section of a function in the 
  # AWS console.
  runtime_version_arn = "arn:aws:lambda:us-east-1::runtime:abcd1234"
}
```

~> Once the runtime update mode is set to `Manual`, the `aws_lambda_function` `runtime` cannot be updated. To upgrade a runtime, the `update_runtime_on` argument must be set to `Auto` or `FunctionUpdate` prior to changing the function's `runtime` argument.

## Argument Reference

The following arguments are required:

* `function_name` - (Required) Name or ARN of the Lambda function.

The following arguments are optional:

* `qualifier` - (Optional) Version of the function. This can be `$LATEST` or a published version number. If omitted, this resource will manage the runtime configuration for `$LATEST`.
* `runtime_version_arn` - (Optional) ARN of the runtime version. Only required when `update_runtime_on` is `Manual`.
* `update_runtime_on` - (Optional) Runtime update mode. Valid values are `Auto`, `FunctionUpdate`, and `Manual`. When a function is created, the default mode is `Auto`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `function_arn` - ARN of the function.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Lambda Runtime Management Config using a comma-delimited string combining `function_name` and `qualifier`. For example:

```terraform
import {
  to = aws_lambda_runtime_management_config.example
  id = "my-function,$LATEST"
}
```

Using `terraform import`, import Lambda Runtime Management Config using a comma-delimited string combining `function_name` and `qualifier`. For example:

```console
% terraform import aws_lambda_runtime_management_config.example my-function,$LATEST
```
