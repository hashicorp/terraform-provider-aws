---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_function_recursion_config"
description: |-
  Terraform resource for managing an AWS Lambda Function Recursion Config.
---

# Resource: aws_lambda_function_recursion_config

Terraform resource for managing an AWS Lambda Function Recursion Config.

## Example Usage

```terraform
resource "aws_lambda_function_recursion_config" "example" {
  function_name  = "testexample"
  recursive_loop = "Allow"
}
```

## Argument Reference

The following arguments are required:

* `function_name` (String) Lambda function name.
* `recursive_loop` (String) Lambda function recursion configuration. Valid values are `Allow` or `Terminate`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` Name of Lambda Function.
* `function_name` Name of Lambda Function.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AWS Lambda Function Recursion Config using the `function_name`. For example:

```terraform
import {
  to = aws_lambda_function_recursion_config.example
  id = "testexample"
}
```

Using `terraform import`, import AWS Lambda Function Recursion Config using the `function_name`. For example:

```console
% terraform import aws_lambda_function_recursion_config.example testexample
```
