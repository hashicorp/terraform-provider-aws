---
subcategory: "Backup"
layout: "aws"
page_title: "AWS: aws_backup_tiering_configuration"
description: |-
  Lists Backup Tiering Configuration resources.
---

# List Resource: aws_backup_tiering_configuration

Lists Backup Tiering Configuration resources.

## Example Usage

```terraform
list "aws_backup_tiering_configuration" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
