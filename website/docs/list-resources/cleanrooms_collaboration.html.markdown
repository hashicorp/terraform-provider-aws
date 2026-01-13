---
subcategory: "Clean Rooms"
layout: "aws"
page_title: "AWS: aws_cleanrooms_collaboration"
description: |-
  Lists Clean Rooms Collaboration resources.
---

# List Resource: aws_cleanrooms_collaboration

Lists Clean Rooms Collaboration resources.

## Example Usage

```terraform
list "aws_cleanrooms_collaboration" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.