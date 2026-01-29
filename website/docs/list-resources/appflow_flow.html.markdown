---
subcategory: "AppFlow"
layout: "aws"
page_title: "AWS: aws_appflow_flow"
description: |-
  Lists AppFlow Flow resources.
---

# List Resource: aws_appflow_flow

Lists AppFlow Flow resources.

## Example Usage

```terraform
list "aws_appflow_flow" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
