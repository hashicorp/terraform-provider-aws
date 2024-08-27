---
subcategory: "Backup"
layout: "aws"
page_title: "AWS: aws_backup_vault"
description: |-
  Provides details about an AWS Backup vault.
---

# Data Source: aws_backup_vault

Use this data source to get information on an existing backup vault.

## Example Usage

```terraform
data "aws_backup_vault" "example" {
  name = "example_backup_vault"
}
```

## Argument Reference

This data source supports the following arguments:

* `name` - (Required) Name of the backup vault.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the vault.
* `kms_key_arn` - Server-side encryption key that is used to protect your backups.
* `recovery_points` - Number of recovery points that are stored in a backup vault.
* `tags` - Metadata that you can assign to help organize the resources that you create.
