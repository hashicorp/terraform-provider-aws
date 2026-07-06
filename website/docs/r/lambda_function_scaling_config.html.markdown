---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_function_scaling_config"
description: |-
  Manages the scaling configuration for an AWS Lambda function.
---

# Resource: aws_lambda_function_scaling_config

Manages the scaling configuration for an AWS Lambda function. The scaling configuration defines the minimum and maximum number of execution environments that can be provisioned for the function, allowing you to control scaling behavior and resource allocation.

~> **NOTE:** This resource only works with Lambda functions that have a capacity provider configuration.

## Example Usage

### Basic Usage

```terraform
resource "aws_lambda_capacity_provider" "example" {
  name = "example"

  vpc_config {
    subnet_ids         = aws_subnet.example[*].id
    security_group_ids = [aws_security_group.example.id]
  }

  permissions_config {
    capacity_provider_operator_role_arn = aws_iam_role.example.arn
  }
}

resource "aws_lambda_function" "example" {
  filename      = "lambda_function.zip"
  function_name = "example"
  role          = aws_iam_role.example.arn
  handler       = "index.handler"
  runtime       = "python3.14"
  memory_size   = 32768
  publish       = true
  publish_to    = "LATEST_PUBLISHED"

  capacity_provider_config {
    lambda_managed_instances_capacity_provider_config {
      capacity_provider_arn = aws_lambda_capacity_provider.example.arn
    }
  }
}

resource "aws_lambda_function_scaling_config" "example" {
  function_name = aws_lambda_function.example.function_name
  qualifier     = "$LATEST.PUBLISHED"

  function_scaling_config {
    min_execution_environments = 3
    max_execution_environments = 100
  }
}
```

## Argument Reference

The following arguments are required:

* `function_name` - (Required) Name or ARN of the Lambda function. Changing this forces a new resource.
* `qualifier` - (Required) Qualifier for the scaling configuration. Valid values: `$LATEST.PUBLISHED` to target the latest published version, or a specific numeric version number (e.g., `1`). Changing this forces a new resource.

The following arguments are optional:

* `function_scaling_config` - (Optional) Scaling configuration block. See [`function_scaling_config` Block](#function_scaling_config-block) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `function_scaling_config` Block

* `max_execution_environments` - (Optional) Maximum number of execution environments that can be provisioned for the function.
* `min_execution_environments` - (Optional) Minimum number of execution environments to maintain for the function.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_lambda_function_scaling_config.example
  identity = {
    function_name = "my-function"
    qualifier     = "$LATEST.PUBLISHED"
  }
}

resource "aws_lambda_function_scaling_config" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `function_name` (String) Name or ARN of the Lambda function.
* `qualifier` (String) Qualifier for the scaling configuration.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Lambda Function Scaling Config using the `function_name` and `qualifier` separated by a comma (`,`). For example:

```terraform
import {
  to = aws_lambda_function_scaling_config.example
  id = "my-function,$LATEST.PUBLISHED"
}
```

Using `terraform import`, import Lambda Function Scaling Config using the `function_name` and `qualifier` separated by a comma (`,`). For example:

```console
% terraform import aws_lambda_function_scaling_config.example my-function,$LATEST.PUBLISHED
```
