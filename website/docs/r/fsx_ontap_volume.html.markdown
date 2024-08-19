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

The following arguments are required:

* `name` - (Required) The name of the Volume. You can use a maximum of 203 alphanumeric characters, plus the underscore (_) special character.
* `storage_virtual_machine_id` - (Required) Specifies the storage virtual machine in which to create the volume.

The following arguments are optional:

* `aggregate_configuration` - (Optional) The Aggregate configuration only applies to `FLEXGROUP` volumes. See [`aggregate_configuration` Block] for details.
* `bypass_snaplock_enterprise_retention` - (Optional) Setting this to `true` allows a SnapLock administrator to delete an FSx for ONTAP SnapLock Enterprise volume with unexpired write once, read many (WORM) files. This configuration must be applied separately before attempting to delete the resource to have the desired behavior. Defaults to `false`.
* `copy_tags_to_backups` - (Optional) A boolean flag indicating whether tags for the volume should be copied to backups. This value defaults to `false`.
* `final_backup_tags` - (Optional) A map of tags to apply to the volume's final backup.
* `junction_path` - (Optional) Specifies the location in the storage virtual machine's namespace where the volume is mounted. The junction_path must have a leading forward slash, such as `/vol3`
* `ontap_volume_type` - (Optional) Specifies the type of volume, valid values are `RW`, `DP`. Default value is `RW`. These can be set by the ONTAP CLI or API. This setting is used as part of migration and replication [Migrating to Amazon FSx for NetApp ONTAP](https://docs.aws.amazon.com/fsx/latest/ONTAPGuide/migrating-fsx-ontap.html)
* `security_style` - (Optional) Specifies the volume security style, Valid values are `UNIX`, `NTFS`, and `MIXED`.
* `size_in_bytes` - (Optional) Specifies the size of the volume, in megabytes (MB), that you are creating. Can be used for any size but required for volumes over 2 PB. Either size_in_bytes or size_in_megabytes must be specified. Minimum size for `FLEXGROUP` volumes are 100GiB per constituent.
* `size_in_megabytes` - (Optional) Specifies the size of the volume, in megabytes (MB), that you are creating. Supported when creating volumes under 2 PB. Either size_in_bytes or size_in_megabytes must be specified. Minimum size for `FLEXGROUP` volumes are 100GiB per constituent.
* `skip_final_backup` - (Optional) When enabled, will skip the default final backup taken when the volume is deleted. This configuration must be applied separately before attempting to delete the resource to have the desired behavior. Defaults to `false`.
* `snaplock_configuration` - (Optional) The SnapLock configuration for an FSx for ONTAP volume. See [`snaplock_configuration` Block](#snaplock_configuration-block) for details.
* `snapshot_policy` - (Optional) Specifies the snapshot policy for the volume. See [snapshot policies](https://docs.aws.amazon.com/fsx/latest/ONTAPGuide/snapshots-ontap.html#snapshot-policies) in the Amazon FSx ONTAP User Guide
* `storage_efficiency_enabled` - (Optional) Set to true to enable deduplication, compression, and compaction storage efficiency features on the volume.
* `tags` - (Optional) A map of tags to assign to the volume. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `tiering_policy` - (Optional) The data tiering policy for an FSx for ONTAP volume. See [`tiering_policy` Block](#tiering_policy-block) for details.
* `volume_style` - (Optional) Specifies the styles of volume, valid values are `FLEXVOL`, `FLEXGROUP`. Default value is `FLEXVOL`. FLEXGROUPS have a larger minimum and maximum size. See Volume Styles for more details. [Volume Styles](https://docs.aws.amazon.com/fsx/latest/ONTAPGuide/volume-styles.html)

### `aggregate_configuration` Block

The `aggregate_configuration` configuration block supports the following arguments:

* `aggregates` - (Optional) Used to specify the names of the aggregates on which the volume will be created. Each aggregate needs to be in the format aggrX where X is the number of the aggregate.
* `constituents_per_aggregate` - (Optional) Used to explicitly set the number of constituents within the FlexGroup per storage aggregate. the default value is `8`.

### `snaplock_configuration` Block

The `snaplock_configuration` configuration block supports the following arguments:

* `snaplock_type` - (Required) Specifies the retention mode of an FSx for ONTAP SnapLock volume. After it is set, it can't be changed. Valid values: `COMPLIANCE`, `ENTERPRISE`.
* `audit_log_volume` - (Optional) Enables or disables the audit log volume for an FSx for ONTAP SnapLock volume. The default value is `false`.
* `autocommit_period` - (Optional) The configuration object for setting the autocommit period of files in an FSx for ONTAP SnapLock volume. See [`autocommit_period` Block](#autocommit_period-block) for details.
* `privileged_delete` - (Optional) Enables, disables, or permanently disables privileged delete on an FSx for ONTAP SnapLock Enterprise volume. Valid values: `DISABLED`, `ENABLED`, `PERMANENTLY_DISABLED`. The default value is `DISABLED`.
* `retention_period` - (Optional) The retention period of an FSx for ONTAP SnapLock volume. See [`retention_period` Block](#retention_period-block) for details.
* `volume_append_mode_enabled` - (Optional) Enables or disables volume-append mode on an FSx for ONTAP SnapLock volume. The default value is `false`.

### `autocommit_period` Block

The `autocommit_period` configuration block supports the following arguments:

* `type` - (Required) The type of time for the autocommit period of a file in an FSx for ONTAP SnapLock volume. Setting this value to `NONE` disables autocommit. Valid values: `MINUTES`, `HOURS`, `DAYS`, `MONTHS`, `YEARS`, `NONE`.
* `value` - (Optional) The amount of time for the autocommit period of a file in an FSx for ONTAP SnapLock volume.

### `retention_period` Block

The `retention_period` configuration block supports the following arguments:

* `default_retention` - (Required) The retention period assigned to a write once, read many (WORM) file by default if an explicit retention period is not set for an FSx for ONTAP SnapLock volume. The default retention period must be greater than or equal to the minimum retention period and less than or equal to the maximum retention period. See [`default_retention` Block](#default_retention-block) for details.
* `maximum_retention` - (Required) The longest retention period that can be assigned to a WORM file on an FSx for ONTAP SnapLock volume. See [`maximum_retention` Block](#maximum_retention-block) for details.
* `minimum_retention` - (Required) The shortest retention period that can be assigned to a WORM file on an FSx for ONTAP SnapLock volume. See [`minimum_retention` Block](#minimum_retention-block) for details.

### `default_retention` Block

The `default_retention` configuration block supports the following arguments:

* `type` - (Required) The type of time for the retention period of an FSx for ONTAP SnapLock volume. Set it to one of the valid types. If you set it to `INFINITE`, the files are retained forever. If you set it to `UNSPECIFIED`, the files are retained until you set an explicit retention period. Valid values: `SECONDS`, `MINUTES`, `HOURS`, `DAYS`, `MONTHS`, `YEARS`, `INFINITE`, `UNSPECIFIED`.
* `value` - (Optional) The amount of time for the autocommit period of a file in an FSx for ONTAP SnapLock volume.

### `maximum_retention` Block

The `maximum_retention` configuration block supports the following arguments:

* `type` - (Required) The type of time for the retention period of an FSx for ONTAP SnapLock volume. Set it to one of the valid types. If you set it to `INFINITE`, the files are retained forever. If you set it to `UNSPECIFIED`, the files are retained until you set an explicit retention period. Valid values: `SECONDS`, `MINUTES`, `HOURS`, `DAYS`, `MONTHS`, `YEARS`, `INFINITE`, `UNSPECIFIED`.
* `value` - (Optional) The amount of time for the autocommit period of a file in an FSx for ONTAP SnapLock volume.

### `minimum_retention` Block

The `minimum_retention` configuration block supports the following arguments:

* `type` - (Required) The type of time for the retention period of an FSx for ONTAP SnapLock volume. Set it to one of the valid types. If you set it to `INFINITE`, the files are retained forever. If you set it to `UNSPECIFIED`, the files are retained until you set an explicit retention period. Valid values: `SECONDS`, `MINUTES`, `HOURS`, `DAYS`, `MONTHS`, `YEARS`, `INFINITE`, `UNSPECIFIED`.
* `value` - (Optional) The amount of time for the autocommit period of a file in an FSx for ONTAP SnapLock volume.

### `tiering_policy` Block

The `tiering_policy` configuration block supports the following arguments:

* `name` - (Required) Specifies the tiering policy for the ONTAP volume for moving data to the capacity pool storage. Valid values are `SNAPSHOT_ONLY`, `AUTO`, `ALL`, `NONE`. Default value is `SNAPSHOT_ONLY`.
* `cooling_period` - (Optional) Specifies the number of days that user data in a volume must remain inactive before it is considered "cold" and moved to the capacity pool. Used with `AUTO` and `SNAPSHOT_ONLY` tiering policies only. Valid values are whole numbers between 2 and 183. Default values are 31 days for `AUTO` and 2 days for `SNAPSHOT_ONLY`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `aggregate_configuration.total_constituents` - The total amount of constituents for a `FLEXGROUP` volume. This would equal constituents_per_aggregate x aggregates.
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
