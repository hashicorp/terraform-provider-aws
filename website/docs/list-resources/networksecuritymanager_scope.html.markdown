---
subcategory: "Network Security Manager"
layout: "aws"
page_title: "AWS: aws_networksecuritymanager_scope"
description: |-
  Lists Network Security Manager Scope resources.
---

# List Resource: aws_networksecuritymanager_scope

Lists Network Security Manager Scope resources.

## Example Usage

```terraform
list "aws_networksecuritymanager_scope" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
