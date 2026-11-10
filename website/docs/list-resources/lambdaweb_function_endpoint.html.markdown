---
subcategory: "Lambda Web"
layout: "aws"
page_title: "AWS: aws_lambdaweb_function_endpoint"
description: |-
  Lists Lambda Web Function Endpoint resources.
---

# List Resource: aws_lambdaweb_function_endpoint

Lists Lambda Web Function Endpoint resources.

## Example Usage

```terraform
list "aws_lambdaweb_function_endpoint" "example" {
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
