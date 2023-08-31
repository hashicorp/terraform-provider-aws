---
subcategory: "FSx"
layout: "aws"
page_title: "AWS: aws_fsx_ontap_volume"
description: |-
  Manages a FSx ONTAP Volume.
---

# Resource: aws_fsx_ontap_volume

Manages a FSx ONTAP Volume.
See the [FSx ONTAP User Guide](https://docs.aws.amazon.com/fsx/latest/ONTAPGuide/managing-volumes.html) for more information.

## Example Usage

### Basic Usage

```terraform
resource "aws_fsx_ontap_volume" "test" {
  name                       = "test"
  junction_path              = "/test"
  size_in_megabytes          = 1024
  storage_efficiency_enabled = true
  storage_virtual_machine_id = aws_fsx_ontap_storage_virtual_machine.test.id
}
```

### Using Tiering Policy

Additional information on tiering policy with ONTAP Volumes can be found in the [FSx ONTAP Guide](https://docs.aws.amazon.com/fsx/latest/ONTAPGuide/managing-volumes.html).

```terraform
resource "aws_fsx_ontap_volume" "test" {
  name                       = "test"
  junction_path              = "/test"
  size_in_megabytes          = 1024
  storage_efficiency_enabled = true
  storage_virtual_machine_id = aws_fsx_ontap_storage_virtual_machine.test.id

  tiering_policy {
    name           = "AUTO"
    cooling_period = 31
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) The name of the Volume. You can use a maximum of 203 alphanumeric characters, plus the underscore (_) special character.
* `junction_path` - (Optional) Specifies the location in the storage virtual machine's namespace where the volume is mounted. The junction_path must have a leading forward slash, such as `/vol3`
* `ontap_volume_type` - (Optional) Specifies the type of volume, valid values are `RW`, `DP`. Default value is `RW`. These can be set by the ONTAP CLI or API. This setting is used as part of migration and replication [Migrating to Amazon FSx for NetApp ONTAP](https://docs.aws.amazon.com/fsx/latest/ONTAPGuide/migrating-fsx-ontap.html)
* `security_style` - (Optional) Specifies the volume security style, Valid values are `UNIX`, `NTFS`, and `MIXED`.
* `size_in_megabytes` - (Required) Specifies the size of the volume, in megabytes (MB), that you are creating.
* `skip_final_backup` - (Optional) When enabled, will skip the default final backup taken when the volume is deleted. This configuration must be applied separately before attempting to delete the resource to have the desired behavior. Defaults to `false`.
* `storage_efficiency_enabled` - (Optional) Set to true to enable deduplication, compression, and compaction storage efficiency features on the volume.
* `storage_virtual_machine_id` - (Required) Specifies the storage virtual machine in which to create the volume.
* `tags` - (Optional) A map of tags to assign to the volume. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### tiering_policy

The `tiering_policy` configuration block supports the following arguments:

* `name` - (Required) Specifies the tiering policy for the ONTAP volume for moving data to the capacity pool storage. Valid values are `SNAPSHOT_ONLY`, `AUTO`, `ALL`, `NONE`. Default value is `SNAPSHOT_ONLY`.
* `cooling_period` - (Optional) Specifies the number of days that user data in a volume must remain inactive before it is considered "cold" and moved to the capacity pool. Used with `AUTO` and `SNAPSHOT_ONLY` tiering policies only. Valid values are whole numbers between 2 and 183. Default values are 31 days for `AUTO` and 2 days for `SNAPSHOT_ONLY`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name of the volune.
* `id` - Identifier of the volume, e.g., `fsvol-12345678`
* `file_system_id` - Describes the file system for the volume, e.g. `fs-12345679`
* `flexcache_endpoint_type` - Specifies the FlexCache endpoint type of the volume, Valid values are `NONE`, `ORIGIN`, `CACHE`. Default value is `NONE`. These can be set by the ONTAP CLI or API and are use with FlexCache feature.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `uuid` - The Volume's UUID (universally unique identifier).
* `volume_type` - The type of volume, currently the only valid value is `ONTAP`.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `delete` - (Default `30m`)
* `update` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import FSx ONTAP volume using the `id`. For example:

```terraform
import {
  to = aws_fsx_ontap_volume.example
  id = "fsvol-12345678abcdef123"
}
```

Using `terraform import`, import FSx ONTAP volume using the `id`. For example:

```console
% terraform import aws_fsx_ontap_volume.example fsvol-12345678abcdef123
```
