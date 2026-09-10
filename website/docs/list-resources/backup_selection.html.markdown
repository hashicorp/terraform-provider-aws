---
subcategory: "Backup"
layout: "aws"
page_title: "AWS: aws_backup_selection"
description: |-
  Lists Backup Selection resources.
---

# List Resource: aws_backup_selection

Lists Backup Selection resources.

## Example Usage

```terraform
list "aws_backup_selection" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
