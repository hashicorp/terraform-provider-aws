---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_functions"
description: |-
  Provides a list of AWS Lambda Functions.
---

# Data Source: aws_lambda_functions

Provides a list of AWS Lambda Functions in the current region. Use this data source to discover existing Lambda functions for inventory, monitoring, or bulk operations.

## Example Usage

### List All Functions

```terraform
data "aws_lambda_functions" "all" {}

output "function_count" {
  value = length(data.aws_lambda_functions.all.function_names)
}

output "all_function_names" {
  value = data.aws_lambda_functions.all.function_names
}
```

### Use Function List for Bulk Operations

```terraform
# Get all Lambda functions
data "aws_lambda_functions" "all" {}

# Create CloudWatch alarms for all functions
resource "aws_cloudwatch_metric_alarm" "lambda_errors" {
  count = length(data.aws_lambda_functions.all.function_names)

  alarm_name          = "${data.aws_lambda_functions.all.function_names[count.index]}-errors"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "Errors"
  namespace           = "AWS/Lambda"
  period              = "300"
  statistic           = "Sum"
  threshold           = "5"
  alarm_description   = "This metric monitors lambda errors"

  dimensions = {
    FunctionName = data.aws_lambda_functions.all.function_names[count.index]
  }

  tags = {
    Environment = "monitoring"
    Purpose     = "lambda-error-tracking"
  }
}
```

### Filter Functions by Name Pattern

```terraform
# Get all functions
data "aws_lambda_functions" "all" {}

# Filter functions with specific naming pattern
locals {
  api_functions = [
    for name in data.aws_lambda_functions.all.function_names :
    name if can(regex("^api-", name))
  ]

  worker_functions = [
    for name in data.aws_lambda_functions.all.function_names :
    name if can(regex("^worker-", name))
  ]
}

output "api_functions" {
  value = local.api_functions
}

output "worker_functions" {
  value = local.worker_functions
}
```

### Create Function Inventory

```terraform
data "aws_lambda_functions" "all" {}

# Get detailed information for each function
data "aws_lambda_function" "details" {
  count         = length(data.aws_lambda_functions.all.function_names)
  function_name = data.aws_lambda_functions.all.function_names[count.index]
}

# Create inventory output
locals {
  function_inventory = [
    for i, name in data.aws_lambda_functions.all.function_names : {
      name        = name
      arn         = data.aws_lambda_functions.all.function_arns[i]
      runtime     = data.aws_lambda_function.details[i].runtime
      memory_size = data.aws_lambda_function.details[i].memory_size
      timeout     = data.aws_lambda_function.details[i].timeout
      handler     = data.aws_lambda_function.details[i].handler
    }
  ]
}

output "function_inventory" {
  value = local.function_inventory
}
```

## Argument Reference

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `function_arns` - List of Lambda Function ARNs.
* `function_names` - List of Lambda Function names.
