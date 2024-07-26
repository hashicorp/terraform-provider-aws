---
subcategory: "Backup"
layout: "aws"
page_title: "AWS: aws_backup_vault"
description: |-
  Provides an AWS Backup vault resource.
---

# Resource: aws_backup_vault

Provides an AWS Backup vault resource.

## Example Usage

```terraform
resource "aws_backup_vault" "example" {
  name        = "example_backup_vault"
  kms_key_arn = aws_kms_key.example.arn
}
```

## Argument Reference

This resource supports the following arguments:

* `force_destroy` - (Optional, Default: `false`) A boolean that indicates that all recovery points stored in the vault are deleted so that the vault can be destroyed without error.
* `kms_key_arn` - (Optional) The server-side encryption key that is used to protect your backups.
* `name` - (Required) Name of the backup vault to create.
* `tags` - (Optional) Metadata that you can assign to help organize the resources that you create. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The name of the vault.
* `arn` - The ARN of the vault.
* `recovery_points` - The number of recovery points that are stored in a backup vault.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Backup vault using the `name`. For example:

```terraform
import {
  to = aws_backup_vault.test-vault
  id = "TestVault"
}
```

Using `terraform import`, import Backup vault using the `name`. For example:

```console
% terraform import aws_backup_vault.test-vault TestVault
```
