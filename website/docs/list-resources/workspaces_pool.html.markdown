---
subcategory: "WorkSpaces"
layout: "aws"
page_title: "AWS: aws_workspaces_pool"
description: |-
  Lists WorkSpaces Pool resources.
---

# List Resource: aws_workspaces_pool

Lists WorkSpaces Pool resources.

## Example Usage

```terraform
list "aws_workspaces_pool" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
