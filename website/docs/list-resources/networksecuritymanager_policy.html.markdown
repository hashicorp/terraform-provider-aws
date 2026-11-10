---
subcategory: "Network Security Manager"
layout: "aws"
page_title: "AWS: aws_networksecuritymanager_policy"
description: |-
  Lists AWS Network Security Manager Policy resources.
---

# List Resource: aws_networksecuritymanager_policy

Lists AWS Network Security Manager Policy resources.

## Example Usage

```terraform
list "aws_networksecuritymanager_policy" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
