---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: aws_ssoadmin_instance"
description: |-
  Lists SSO Admin Instance resources.
---

# List Resource: aws_ssoadmin_instance

Lists SSO Admin Instance resources.

## Example Usage

```terraform
list "aws_ssoadmin_instance" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
