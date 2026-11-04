---
subcategory: "RAM (Resource Access Manager)"
layout: "aws"
page_title: "AWS: aws_ram_resource_association"
description: |-
  Lists RAM (Resource Access Manager) Resource Association resources.
---

# List Resource: aws_ram_resource_association

Lists RAM (Resource Access Manager) Resource Association resources.

## Example Usage

```terraform
list "aws_ram_resource_association" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
