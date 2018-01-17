---
layout: "aws"
page_title: "AWS: aws_lambda_function"
sidebar_current: "docs-aws-datasource-lambda-function"
description: |-
    Provides details about a specific Lambda function.
---

# aws_lambda_function

Provides details about a specific Lambda function.

This resource may prove useful when functions are built with external libraries or
otherwise managed outside of Terraform.

## Example Usage

variable "function_name" {
  type = "string"
}

variable "function_alias" {
  type = "string"
}

data "aws_lambda_function" "function" {
  function_name = "${var.function_name}"
  version       = "${var.function_alias}"
}

## Argument Reference

The following arguments are supported:

* `function_name` - (Required) Either the short name of the function or the _unqualified_ function ARN (ARN without version number or alias).
* `qualifier` - (Optional) The version or alias of the function. Defaults to `$LATEST`.

## Attribute Reference

The following attributes are exported:

* `arn` - The fully-qualified ARN of the function.
* `function_name` - The "short" function name.
* `role` - The ARN of the IAM role that the function assumes on execution.
* `version` - The version of the function. This will be set to the actual version number if `qualifier` argument was an alias. If no qualifier was set this will be set to `$LATEST`.
