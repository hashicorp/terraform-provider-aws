---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_function"
description: |-
  Lists Lambda Function resources.
---

# List Resource: aws_lambda_function

Lists Lambda Function resources.

## Example Usage

```terraform
list "aws_lambda_function" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
