---
subcategory: "Backup"
layout: "aws"
page_title: "AWS: aws_backup_plan"
description: |-
  Lists Backup Plan resources.
---

# List Resource: aws_backup_plan

Lists Backup Plan resources.

## Example Usage

```terraform
list "aws_backup_plan" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
