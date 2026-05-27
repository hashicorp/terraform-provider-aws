---
subcategory: "Clean Rooms"
layout: "aws"
page_title: "AWS: aws_cleanrooms_configured_table"
description: |-
  Lists Clean Rooms Configured Table resources.
---

# List Resource: aws_cleanrooms_configured_table

Lists Clean Rooms Configured Table resources.

## Example Usage

```terraform
list "aws_cleanrooms_configured_table" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
