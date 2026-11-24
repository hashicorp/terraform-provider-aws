---
subcategory: "Network Security Manager"
layout: "aws"
page_title: "AWS: aws_networksecuritymanager_template"
description: |-
  Lists AWS Network Security Manager Template resources.
---

# List Resource: aws_networksecuritymanager_template

Lists AWS Network Security Manager Template resources.

## Example Usage

```terraform
list "aws_networksecuritymanager_template" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
