---
subcategory: "Oracle Database@AWS"
layout: "aws"
page_title: "AWS: aws_odb_exascale_db_storage_vault"
description: |-
  Manages an Oracle Database@AWS Exascale DB storage vault.
---

# Resource: aws_odb_exascale_db_storage_vault

Manages an [Oracle Database@AWS Exascale DB storage vault](https://docs.aws.amazon.com/odb/latest/APIReference/API_CreateExascaleDbStorageVault.html).

## Example Usage

### Basic Usage

```terraform
resource "aws_odb_exascale_db_storage_vault" "example" {
  availability_zone_id                             = "use1-az2"
  display_name                                     = "example"
  high_capacity_database_storage_total_size_in_gbs = 300
}
```

### Autoscaling

```terraform
resource "aws_odb_exascale_db_storage_vault" "example" {
  autoscale_limit_in_gbs                           = 600
  availability_zone_id                             = "use1-az2"
  display_name                                     = "example"
  high_capacity_database_storage_total_size_in_gbs = 300
  is_autoscale_enabled                             = true
}
```

## Argument Reference

The following arguments are required:

* `availability_zone_id` - (Required) Availability Zone ID for the Exascale DB storage vault. Must be between `1` and `255` characters. Changing this value creates a new resource.
* `display_name` - (Required) User-friendly name for the Exascale DB storage vault. Must be between `1` and `255` characters, start with a letter or underscore, contain only letters, numbers, underscores, or hyphens, and not contain consecutive hyphens.
* `high_capacity_database_storage_total_size_in_gbs` - (Required) Total size of the high-capacity database storage, in GB. Must be `0` or greater.

The following arguments are optional:

* `additional_flash_cache_in_percent` - (Optional) Additional flash cache percentage for the Exascale DB storage vault. Must be `0` or greater.
* `autoscale_limit_in_gbs` - (Optional) Autoscale limit for the Exascale DB storage vault, in GB. Must be `0` or greater.
* `availability_zone` - (Optional) Availability Zone for the Exascale DB storage vault. Must be between `1` and `255` characters. Changing this value creates a new resource.
* `description` - (Optional) Description of the Exascale DB storage vault. Must be between `1` and `400` characters.
* `is_autoscale_enabled` - (Optional) Whether autoscaling is enabled for the Exascale DB storage vault.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `time_zone` - (Optional) Time zone for the Exascale DB storage vault. Must be between `1` and `255` characters. Changing this value creates a new resource.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Exascale DB storage vault.
* `id` - Unique identifier of the Exascale DB storage vault.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `24h`)
* `update` - (Default `24h`)
* `delete` - (Default `24h`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_odb_exascale_db_storage_vault.example
  identity = {
    id = "xsvault_0123456789"
  }
}

resource "aws_odb_exascale_db_storage_vault" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `id` (String) Exascale DB storage vault ID.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an Exascale DB storage vault using `id`. For example:

```terraform
import {
  to = aws_odb_exascale_db_storage_vault.example
  id = "xsvault_0123456789"
}
```

Using `terraform import`, import an Exascale DB storage vault using `id`. For example:

```console
% terraform import aws_odb_exascale_db_storage_vault.example xsvault_0123456789
```
