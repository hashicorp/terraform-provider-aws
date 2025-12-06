---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_capacity_provider"
description: |-
  Lists Lambda Capacity Provider resources.
---

# List Resource: aws_lambda_capacity_provider

Lists Lambda Capacity Provider resources.

## Example Usage

```terraform
list "aws_lambda_capacity_provider" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
