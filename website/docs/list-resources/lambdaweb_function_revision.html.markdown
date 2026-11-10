---
subcategory: "Lambda Web"
layout: "aws"
page_title: "AWS: aws_lambdaweb_function_revision"
description: |-
  Lists Lambda Web Function Revision resources.
---

# List Resource: aws_lambdaweb_function_revision

Lists Lambda Web Function Revision resources.

## Example Usage

```terraform
list "aws_lambdaweb_function_revision" "example" {
  provider = aws

  config {
    function_name = aws_lambdaweb_function.example.function_name
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `function_name` - (Required) Web function name.
* `region` - (Optional) Region to query. Defaults to provider region.
