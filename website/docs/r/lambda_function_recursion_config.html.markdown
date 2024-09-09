---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_function_recursion_config"
description: |-
  Terraform resource for managing an AWS Lambda Function Recursion Config.
---

# Resource: aws_lambda_function_recursion_config

Terraform resource for managing an AWS Lambda Function Recursion Config.

~> Destruction of this resource will return the `recursive_loop` configuration back to the default value of `Terminate`.

## Example Usage

```terraform
resource "aws_lambda_function_recursion_config" "example" {
  function_name  = "SomeFunction"
  recursive_loop = "Allow"
}
```

## Argument Reference

The following arguments are required:

* `function_name` - (Required) Lambda function name.
* `recursive_loop` - (Required) Lambda function recursion configuration. Valid values are `Allow` or `Terminate`.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AWS Lambda Function Recursion Config using the `function_name`. For example:

```terraform
import {
  to = aws_lambda_function_recursion_config.example
  id = "SomeFunction"
}
```

Using `terraform import`, import AWS Lambda Function Recursion Config using the `function_name`. For example:

```console
% terraform import aws_lambda_function_recursion_config.example SomeFunction
```
