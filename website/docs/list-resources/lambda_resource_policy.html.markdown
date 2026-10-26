---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_resource_policy"
description: |-
  Lists Lambda Resource Policy resources.
---

# List Resource: aws_lambda_resource_policy

Lists Lambda Resource Policy resources.

## Example Usage

```terraform
list "aws_lambda_resource_policy" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
