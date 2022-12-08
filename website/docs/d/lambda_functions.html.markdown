---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_functions"
description: |-
  Provides a Lambda Functions data source.
---

# aws_lambda_function

Terraform data resource to get a list of Lambda Functions for a provider.

## Example Usage

```terraform
data "aws_lambda_functions" "all" {}
```

## Argument Reference

The resource does not support any arguments.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `function_names` - A list of Lambda Function names.
* `function_arns` - A list of Lambda Function ARNs.