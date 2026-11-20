---
subcategory: "Network Security Manager"
layout: "aws"
page_title: "AWS: aws_networksecuritymanager_rule"
description: |-
  Lists AWS Network Security Manager Rule resources.
---

# List Resource: aws_networksecuritymanager_rule

Lists AWS Network Security Manager Rule resources.

## Example Usage

```terraform
list "aws_networksecuritymanager_rule" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
