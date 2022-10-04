---
subcategory: "S3 Control"
layout: "aws"
page_title: "AWS: aws_s3control_storage_lens_configuration"
description: |-
  Provides a resource to manage an S3 Storage Lens configuration.
---

# Resource: aws_s3control_storage_lens_configuration

Provides a resource to manage an S3 Storage Lens configuration.

## Example Usage

```terraform
resource "aws_s3control_storage_lens_configuration" "example" {
  config_id = "example-1"

  storage_lens_configuration {
    enabled = true

    account_level {
      bucket_level {}
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `account_id` - (Optional) The AWS account ID for the S3 Storage Lens configuration. Defaults to automatically determined account ID of the Terraform AWS provider.
* `config_id` - (Required) The ID of the S3 Storage Lens configuration.
* `storage_lens_configuration` - (Required) The S3 Storage Lens configuration. See [Storage Lens Configuration](#storage-lens-configuration) below for more details.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Storage Lens Configuration

The `storage_lens_configuration` block supports the following:

* `account_level` (Required) The account-level configurations of the S3 Storage Lens configuration. See [Account Level](#account-level) below for more details.
* `enabled` (Required) Whether the S3 Storage Lens configuration is enabled.

### Account Level

The `account_level` block supports the following:

* `activity_metrics` (Optional) S3 Storage Lens activity metrics. See [Activity Metrics](#activity-metrics) below for more details.
* `bucket_level` (Required) S3 Storage Lens bucket-level configuration. See [Bucket Level](#bucket-level) below for more details.

### Activity Metrics

The `activity_metrics` block supports the following:

* `enabled` (Optional) Whether the activity metrics are enabled.

### Bucket Level

The `bucket_level` block supports the following:

* `activity_metrics` (Optional) S3 Storage Lens activity metrics. See [Activity Metrics](#activity-metrics) above for more details.
* `prefix_level` (Optional) Prefix-level metrics for S3 Storage Lens. See [Prefix Level](#prefix-level) below for more details.

### Prefix Level

The `prefix_level` block supports the following:

* `storage_metrics` (Optional) Prefix-level storage metrics for S3 Storage Lens. See [Prefix Level Storage Metrics](#prefix-level-storage-metrics) below for more details.

### Prefix Level Storage Metrics

The `storage_metrics` block supports the following:

* `enabled` (Optional) Whether prefix-level storage metrics are enabled.
* `selection_criteria` (Optional) Selection criteria. See [Selection Criteria](#selection-criteria) below for more details.

### Selection Criteria

The `selection_criteria` block supports the following:

* `delimiter` (Optional) The delimiter of the selection criteria being used.
* `max_depth` (Optional) The max depth of the selection criteria.
* `min_storage_bytes_percentage` (Optional) The minimum number of storage bytes percentage whose metrics will be selected.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the S3 Storage Lens configuration.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

S3 Storage Lens configurations can be imported using the `account_id` and `config_id`, separated by a colon (`:`), e.g.

```
$ terraform import aws_s3control_storage_lens_configuration.example 123456789012:example-1
```