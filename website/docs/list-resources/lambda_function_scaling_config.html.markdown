---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_function_scaling_config"
description: |-
  Lists Lambda Function Scaling Config resources.
---

# List Resource: aws_lambda_function_scaling_config

Lists Lambda Function Scaling Config resources. Scaling configurations apply to function versions that use a capacity provider, so this enumerates the function versions attached to each capacity provider and returns those that have a scaling configuration.

## Example Usage

```terraform
list "aws_lambda_function_scaling_config" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
