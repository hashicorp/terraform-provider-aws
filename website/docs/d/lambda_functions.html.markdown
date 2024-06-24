---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_functions"
description: |-
  Terraform data resource to get a list of Lambda Functions.
---

# aws_lambda_functions

Terraform data resource to get a list of Lambda Functions.

## Example Usage

```terraform
data "aws_lambda_functions" "all" {}
```

## Argument Reference

The resource does not support any arguments.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `function_names` - A list of Lambda Function names.
* `function_arns` - A list of Lambda Function ARNs.
