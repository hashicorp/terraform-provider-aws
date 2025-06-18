---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_invocation"
description: |-
  Invokes an AWS Lambda Function and returns its results.
---

# Data Source: aws_lambda_invocation

Invokes an AWS Lambda Function and returns its results. Use this data source to execute Lambda functions during Terraform operations and use their results in other resources or outputs.

The Lambda function is invoked with [RequestResponse](https://docs.aws.amazon.com/lambda/latest/dg/API_Invoke.html#API_Invoke_RequestSyntax) invocation type.

~> **Note:** The `aws_lambda_invocation` data source invokes the function during the first `apply` and every subsequent `plan` when the function is known.

~> **Note:** If you get a `KMSAccessDeniedException: Lambda was unable to decrypt the environment variables because KMS access was denied` error when invoking a Lambda function with environment variables, the IAM role associated with the function may have been deleted and recreated after the function was created. You can fix the problem two ways: 1) updating the function's role to another role and then updating it back again to the recreated role, or 2) by using Terraform to `taint` the function and `apply` your configuration again to recreate the function. (When you create a function, Lambda grants permissions on the KMS key to the function's IAM role. If the IAM role is recreated, the grant is no longer valid. Changing the function's role or recreating the function causes Lambda to update the grant.)

## Example Usage

### Basic Invocation

```terraform
data "aws_lambda_invocation" "example" {
  function_name = aws_lambda_function.example.function_name

  input = jsonencode({
    operation = "getStatus"
    id        = "123456"
  })
}

output "result" {
  value = jsondecode(data.aws_lambda_invocation.example.result)
}
```

### Dynamic Resource Configuration

```terraform
# Get resource configuration from Lambda
data "aws_lambda_invocation" "resource_config" {
  function_name = "resource-config-generator"
  qualifier     = "production" # Use production alias

  input = jsonencode({
    environment = var.environment
    region      = data.aws_region.current.name
    service     = "api"
  })
}

locals {
  config = jsondecode(data.aws_lambda_invocation.resource_config.result)
}

# Use dynamic configuration
resource "aws_elasticache_cluster" "example" {
  cluster_id           = local.config.cache.cluster_id
  engine               = local.config.cache.engine
  node_type            = local.config.cache.node_type
  num_cache_nodes      = local.config.cache.nodes
  parameter_group_name = local.config.cache.parameter_group

  tags = local.config.tags
}
```

### Error Handling

```terraform
data "aws_lambda_invocation" "example" {
  function_name = aws_lambda_function.example.function_name

  input = jsonencode({
    action  = "validate"
    payload = var.configuration
  })
}

locals {
  result = jsondecode(data.aws_lambda_invocation.example.result)

  # Check for errors in the response
  has_errors     = try(local.result.errors != null, false)
  error_messages = local.has_errors ? join(", ", local.result.errors) : null
}

# Fail the apply if validation fails
resource "null_resource" "validation_check" {
  count = local.has_errors ? fail("Configuration validation failed: ${local.error_messages}") : 0
}
```

## Argument Reference

The following arguments are required:

* `function_name` - (Required) Name of the Lambda function.
* `input` - (Required) String in JSON format that is passed as payload to the Lambda function.

The following arguments are optional:

* `qualifier` - (Optional) Qualifier (a.k.a version) of the Lambda function. Defaults to `$LATEST`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `result` - String result of the Lambda function invocation.
