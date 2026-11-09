---
subcategory: "Backup"
layout: "aws"
page_title: "AWS: aws_backup_vault"
description: |-
  Lists Backup Vault resources.
---

# List Resource: aws_backup_vault

Lists Backup Vault resources.

## Example Usage

```terraform
list "aws_backup_vault" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
