---
subcategory: "Backup"
layout: "aws"
page_title: "AWS: aws_backup_tiering_configuration"
description: |-
  Provides an AWS Backup tiering configuration resource.
---

# Resource: aws_backup_tiering_configuration

Provides an AWS Backup tiering configuration resource. A tiering configuration allows you to configure automatic transitions of backup data from the warm storage tier to the low-cost cold storage tier for cost optimization.

## Example Usage

### Basic Usage

```terraform
resource "aws_backup_vault" "example" {
  name = "example"
}

resource "aws_backup_tiering_configuration" "example" {
  tiering_configuration_name = "example_tiering_config"
  backup_vault_name          = aws_backup_vault.example.name

  resource_selection {
    resource_type                 = "S3"
    resources                     = ["*"]
    tiering_down_settings_in_days = 90
  }
}
```

### All Backup Vaults with Wildcard

```terraform
resource "aws_backup_vault" "example" {
  name = "example"
}

resource "aws_backup_tiering_configuration" "example" {
  tiering_configuration_name = "all_vaults_tiering"
  backup_vault_name          = "*"

  tags = {
    Environment = "production"
    Team        = "backup"
  }

  resource_selection {
    resource_type                 = "S3"
    resources                     = ["*"]
    tiering_down_settings_in_days = 180
  }
}
```

### Multiple Resource Selections

```terraform
resource "aws_backup_vault" "example" {
  name = "example"
}

resource "aws_backup_tiering_configuration" "example" {
  tiering_configuration_name = "example_tiering_config"
  backup_vault_name          = aws_backup_vault.example.name

  resource_selection {
    resource_type                 = "S3"
    resources                     = ["arn:aws:s3:::specific-bucket-1/*"]
    tiering_down_settings_in_days = 90
  }

  resource_selection {
    resource_type                 = "S3"
    resources                     = ["arn:aws:s3:::specific-bucket-2/*"]
    tiering_down_settings_in_days = 60
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tiering_configuration_name` - (Required) The unique name of the tiering configuration. This cannot be changed after creation, and it must consist of only alphanumeric characters and underscores. Maximum length is 200 characters.
* `backup_vault_name` - (Required) The name of the backup vault where the tiering configuration applies. Use `*` to apply to all backup vaults.
* `resource_selection` - (Required) An array of resource selection objects that specify which resources are included in the tiering configuration and their tiering settings. Maximum of 5 resource selections per tiering configuration. See [Resource Selection](#resource-selection) below.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider `default_tags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.

### Resource Selection

`resource_selection` supports the following attributes:

* `resource_type` - (Required) The type of AWS resource. Currently limited to `S3`.
* `resources` - (Required) An array of strings that either contains ARNs of the associated resources or contains a wildcard `*` to specify all resources. You can specify up to 100 specific resources per tiering configuration.
* `tiering_down_settings_in_days` - (Required) The number of days after creation within a backup vault that an object can transition to the low cost warm storage tier. Must be a positive integer between 60 and 36500 days.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `tiering_configuration_arn` - The Amazon Resource Name (ARN) that uniquely identifies the tiering configuration.
* `creation_time` - The date and time the tiering configuration was created, in UTC format.
* `last_updated_time` - The date and time the tiering configuration was updated, in UTC format.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider `default_tags` configuration block.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Backup Tiering Configuration using the `tiering_configuration_name`. For example:

```terraform
import {
  to = aws_backup_tiering_configuration.example
  id = "example_tiering_config"
}
```

Using `terraform import`, import Backup Tiering Configuration using the `name`. For example:

```console
% terraform import aws_backup_tiering_configuration.example example_tiering_config
```
