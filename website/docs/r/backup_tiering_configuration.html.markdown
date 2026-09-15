---
subcategory: "Backup"
layout: "aws"
page_title: "AWS: aws_backup_tiering_configuration"
description: |-
  Manages an AWS Backup Tiering Configuration.
---

# Resource: aws_backup_tiering_configuration

Manages an AWS Backup Tiering Configuration. A tiering configuration moves S3 backups in a backup vault to the low-cost warm storage tier after a configurable number of days. See [Tiering S3 backups](https://docs.aws.amazon.com/aws-backup/latest/devguide/s3-backups.html) in the AWS Backup Developer Guide for details.

## Example Usage

### Basic Usage

```terraform
resource "aws_backup_vault" "example" {
  name = "example-vault"
}

resource "aws_backup_tiering_configuration" "example" {
  name              = "example_tiering_configuration"
  backup_vault_name = aws_backup_vault.example.name

  resource_selection {
    resource_type                 = "S3"
    resources                     = ["*"]
    tiering_down_settings_in_days = 60
  }
}
```

### All Backup Vaults

```terraform
resource "aws_backup_tiering_configuration" "example" {
  name              = "example_tiering_configuration"
  backup_vault_name = "*"

  resource_selection {
    resource_type                 = "S3"
    resources                     = ["*"]
    tiering_down_settings_in_days = 60
  }
}
```

## Argument Reference

The following arguments are required:

* `backup_vault_name` - (Required) Name of the backup vault the tiering configuration applies to. Use `*` to apply the configuration to all backup vaults, in which case every `resource_selection` block must select all resources (`resources = ["*"]`).
* `name` - (Required) Name of the tiering configuration. Must contain only alphanumeric characters and underscores, up to 200 characters. Changing this forces a new resource.
* `resource_selection` - (Required) Resources included in the tiering configuration and their tiering settings. Up to 5 blocks are allowed. See [`resource_selection` Block](#resource_selection-block) below.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `resource_selection` Block

The `resource_selection` block supports the following arguments:

* `resource_type` - (Required) Type of AWS resource. Valid value is `S3`. See [Feature availability by resource](https://docs.aws.amazon.com/aws-backup/latest/devguide/backup-feature-availability.html#features-by-resource) in the AWS Backup Developer Guide for supported resource types.
* `resources` - (Required) Set of ARNs of the resources to include, or `*` to include all resources of the given type. A tiering configuration can include up to 100 ARNs across all `resource_selection` blocks, each ARN can appear in only one block, and `*` cannot be combined with specific ARNs.
* `tiering_down_settings_in_days` - (Required) Number of days after creation within the backup vault before an object can transition to the low-cost warm storage tier. Valid value is between `60` and `36500`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the tiering configuration.
* `creation_time` - Date and time the tiering configuration was created, in RFC3339 format.
* `last_updated_time` - Date and time the tiering configuration was last updated, in RFC3339 format.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_backup_tiering_configuration.example
  identity = {
    name = "example_tiering_configuration"
  }
}

resource "aws_backup_tiering_configuration" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `name` (String) Name of the tiering configuration.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Backup Tiering Configuration using the `name`. For example:

```terraform
import {
  to = aws_backup_tiering_configuration.example
  id = "example_tiering_configuration"
}
```

Using `terraform import`, import Backup Tiering Configuration using the `name`. For example:

```console
% terraform import aws_backup_tiering_configuration.example example_tiering_configuration
```
