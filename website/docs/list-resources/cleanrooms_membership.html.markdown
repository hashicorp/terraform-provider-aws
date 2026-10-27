---
subcategory: "Clean Rooms"
layout: "aws"
page_title: "AWS: aws_cleanrooms_membership"
description: |-
  Lists Clean Rooms Membership resources.
---

# List Resource: aws_cleanrooms_membership

Lists Clean Rooms Membership resources.

## Example Usage

```terraform
list "aws_cleanrooms_membership" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
