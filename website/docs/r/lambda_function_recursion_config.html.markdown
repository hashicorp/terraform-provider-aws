---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_function_recursion_config"
description: |-
  Manages an AWS Lambda Function Recursion Config.
---

# Resource: aws_lambda_function_recursion_config

Manages an AWS Lambda Function Recursion Config. Use this resource to control how Lambda handles recursive function invocations to prevent infinite loops.

~> **Note:** Destruction of this resource will return the `recursive_loop` configuration back to the default value of `Terminate`.

## Example Usage

### Allow Recursive Invocations

```terraform
# Lambda function that may need to call itself
resource "aws_lambda_function" "example" {
  filename      = "function.zip"
  function_name = "recursive_processor"
  role          = aws_iam_role.lambda_role.arn
  handler       = "index.handler"
  runtime       = "python3.12"
}

# Allow the function to invoke itself recursively
resource "aws_lambda_function_recursion_config" "example" {
  function_name  = aws_lambda_function.example.function_name
  recursive_loop = "Allow"
}
```

### Production Safety Configuration

```terraform
# Production function with recursion protection
resource "aws_lambda_function" "production_processor" {
  filename      = "processor.zip"
  function_name = "production-data-processor"
  role          = aws_iam_role.lambda_role.arn
  handler       = "app.handler"
  runtime       = "nodejs20.x"

  tags = {
    Environment = "production"
    Purpose     = "data-processing"
  }
}

# Prevent infinite loops in production
resource "aws_lambda_function_recursion_config" "example" {
  function_name  = aws_lambda_function.production_processor.function_name
  recursive_loop = "Terminate" # Safety first in production
}
```

## Argument Reference

The following arguments are required:

* `function_name` - (Required) Name of the Lambda function.
* `recursive_loop` - (Required) Lambda function recursion configuration. Valid values are `Allow` or `Terminate`.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Lambda Function Recursion Config using the `function_name`. For example:

```terraform
import {
  to = aws_lambda_function_recursion_config.example
  id = "recursive_processor"
}
```

For backwards compatibility, the following legacy `terraform import` command is also supported:

```console
% terraform import aws_lambda_function_recursion_config.example recursive_processor
```
