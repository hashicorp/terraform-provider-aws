---
subcategory: "Backup"
layout: "aws"
page_title: "AWS: aws_backup_vault_lock_configuration"
description: |-
  Provides an AWS Backup vault lock configuration resource.
---

# Resource: aws_backup_vault_lock_configuration

Provides an AWS Backup vault lock configuration resource.

## Example Usage

```terraform
resource "aws_backup_vault_lock_configuration" "test" {
  backup_vault_name   = "example_backup_vault"
  changeable_for_days = 3
  max_retention_days  = 1200
  min_retention_days  = 7
}
```

## Argument Reference

This resource supports the following arguments:

* `backup_vault_name` - (Required) Name of the backup vault to add a lock configuration for.
* `changeable_for_days` - (Optional) The number of days before the lock date. If omitted creates a vault lock in `governance` mode, otherwise it will create a vault lock in `compliance` mode.
* `max_retention_days` - (Optional) The maximum retention period that the vault retains its recovery points.
* `min_retention_days` - (Optional) The minimum retention period that the vault retains its recovery points.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `backup_vault_name` - The name of the vault.
* `backup_vault_arn` - The ARN of the vault.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Backup vault lock configuration using the `name`. For example:

```terraform
import {
  to = aws_backup_vault_lock_configuration.test
  id = "TestVault"
}
```

Using `terraform import`, import Backup vault lock configuration using the `name`. For example:

```console
% terraform import aws_backup_vault_lock_configuration.test TestVault
```
