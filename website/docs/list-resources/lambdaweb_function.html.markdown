---
subcategory: "Lambda Web"
layout: "aws"
page_title: "AWS: aws_lambdaweb_function"
description: |-
  Lists Lambda Web Function resources.
---

# List Resource: aws_lambdaweb_function

Lists Lambda Web Function resources.

## Example Usage

```terraform
list "aws_lambdaweb_function" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
