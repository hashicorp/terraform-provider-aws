---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_layer_version"
description: |-
  Lists Lambda Layer Version resources.
---

# List Resource: aws_lambda_layer_version

Lists Lambda Layer Version resources.

## Example Usage

```terraform
list "aws_lambda_layer_version" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
