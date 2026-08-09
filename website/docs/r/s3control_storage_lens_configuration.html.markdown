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
data "aws_caller_identity" "current" {}

resource "aws_s3control_storage_lens_configuration" "example" {
  config_id = "example-1"

  storage_lens_configuration {
    enabled = true

    account_level {
      activity_metrics {
        enabled = true
      }

      bucket_level {
        activity_metrics {
          enabled = true
        }
      }
    }

    data_export {
      cloud_watch_metrics {
        enabled = true
      }

      s3_bucket_destination {
        account_id            = data.aws_caller_identity.current.account_id
        arn                   = aws_s3_bucket.target.arn
        format                = "CSV"
        output_schema_version = "V_1"

        encryption {
          sse_s3 {}
        }
      }
    }

    exclude {
      buckets = [aws_s3_bucket.b1.arn, aws_s3_bucket.b2.arn]
      regions = ["us-east-2"]
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `account_id` - (Optional) AWS account ID for the S3 Storage Lens configuration. Defaults to automatically determined account ID of the Terraform AWS provider.
* `config_id` - (Required) ID of the S3 Storage Lens configuration.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `storage_lens_configuration` - (Required) S3 Storage Lens configuration. See [`storage_lens_configuration`](#storage_lens_configuration-block) below for more details.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `storage_lens_configuration` Block

The `storage_lens_configuration` block supports the following:

* `account_level` - (Required) Account-level configurations of the S3 Storage Lens configuration. See [`account_level`](#account_level-block) below for more details.
* `aws_org` - (Optional) Amazon Web Services organization for the S3 Storage Lens configuration. See [`aws_org`](#aws_org-block) below for more details.
* `data_export` - (Optional) Properties of S3 Storage Lens metrics export including the destination, schema and format. See [`data_export`](#data_export-block) below for more details.
* `enabled` - (Required) Whether the S3 Storage Lens configuration is enabled.
* `exclude` - (Optional) What is excluded in this configuration. Conflicts with `include`. See [`exclude`](#exclude-block) below for more details.
* `expanded_prefixes_data_export` - (Optional) Configuration for the S3 Storage Lens expanded prefix metrics report. Unlike the default Storage Lens metrics report, the enhanced prefix metrics report includes all S3 Storage Lens storage and activity data related to the full list of prefixes in your Storage Lens configuration. See [`expanded_prefixes_data_export`](#expanded_prefixes_data_export-block) below for more details.
* `include` - (Optional) What is included in this configuration. Conflicts with `exclude`. See [`include`](#include-block) below for more details.
* `prefix_delimiter` - (Optional) Prefix delimiter used for object keys in this S3 Storage Lens configuration.

### `account_level` Block

The `account_level` block supports the following:

* `activity_metrics` - (Optional) S3 Storage Lens activity metrics. See [`activity_metrics`](#activity_metrics-block) below for more details.
* `advanced_cost_optimization_metrics` - (Optional) Advanced cost-optimization metrics for S3 Storage Lens. See [`advanced_cost_optimization_metrics`](#advanced_cost_optimization_metrics-block) below for more details.
* `advanced_data_protection_metrics` - (Optional) Advanced data-protection metrics for S3 Storage Lens. See [`advanced_data_protection_metrics`](#advanced_data_protection_metrics-block) below for more details.
* `advanced_performance_metrics` - (Optional) Advanced performance metrics for S3 Storage Lens. See [`advanced_performance_metrics`](#advanced_performance_metrics-block) below for more details.
* `bucket_level` - (Required) S3 Storage Lens bucket-level configuration. See [`bucket_level`](#bucket_level-block) below for more details.
* `detailed_status_code_metrics` - (Optional) Detailed status code metrics for S3 Storage Lens. See [`detailed_status_code_metrics`](#detailed_status_code_metrics-block) below for more details.

### `activity_metrics` Block

The `activity_metrics` block supports the following:

* `enabled` - (Optional) Whether the activity metrics are enabled.

### `advanced_cost_optimization_metrics` Block

The `advanced_cost_optimization_metrics` block supports the following:

* `enabled` - (Optional) Whether advanced cost-optimization metrics are enabled.

### `advanced_data_protection_metrics` Block

The `advanced_data_protection_metrics` block supports the following:

* `enabled` - (Optional) Whether advanced data-protection metrics are enabled.

### `advanced_performance_metrics` Block

The `advanced_performance_metrics` block supports the following:

* `enabled` - (Optional) Whether advanced performance metrics are enabled.

### `detailed_status_code_metrics` Block

The `detailed_status_code_metrics` block supports the following:

* `enabled` - (Optional) Whether detailed status code metrics are enabled.

### `bucket_level` Block

The `bucket_level` block supports the following:

* `activity_metrics` - (Optional) S3 Storage Lens activity metrics. See [`activity_metrics`](#activity_metrics-block) above for more details.
* `advanced_cost_optimization_metrics` - (Optional) Advanced cost-optimization metrics for S3 Storage Lens. See [`advanced_cost_optimization_metrics`](#advanced_cost_optimization_metrics-block) above for more details.
* `advanced_data_protection_metrics` - (Optional) Advanced data-protection metrics for S3 Storage Lens. See [`advanced_data_protection_metrics`](#advanced_data_protection_metrics-block) above for more details.
* `advanced_performance_metrics` - (Optional) Advanced performance metrics for S3 Storage Lens. See [`advanced_performance_metrics`](#advanced_performance_metrics-block) above for more details.
* `detailed_status_code_metrics` - (Optional) Detailed status code metrics for S3 Storage Lens. See [`detailed_status_code_metrics`](#detailed_status_code_metrics-block) above for more details.
* `prefix_level` - (Optional) Prefix-level metrics for S3 Storage Lens. See [`prefix_level`](#prefix_level-block) below for more details.

### `prefix_level` Block

The `prefix_level` block supports the following:

* `storage_metrics` - (Required) Prefix-level storage metrics for S3 Storage Lens. See [`storage_metrics`](#storage_metrics-block) below for more details.

### `storage_metrics` Block

The `storage_metrics` block supports the following:

* `enabled` - (Optional) Whether prefix-level storage metrics are enabled.
* `selection_criteria` - (Optional) Selection criteria. See [`selection_criteria`](#selection_criteria-block) below for more details.

### `selection_criteria` Block

The `selection_criteria` block supports the following:

* `delimiter` - (Optional) Delimiter of the selection criteria being used.
* `max_depth` - (Optional) Max depth of the selection criteria.
* `min_storage_bytes_percentage` - (Optional) Minimum number of storage bytes percentage whose metrics will be selected.

### `aws_org` Block

The `aws_org` block supports the following:

* `arn` - (Required) Amazon Resource Name (ARN) of the Amazon Web Services organization.

### `data_export` Block

The `data_export` block supports the following:

* `cloud_watch_metrics` - (Optional) Amazon CloudWatch publishing for S3 Storage Lens metrics. See [`cloud_watch_metrics`](#cloud_watch_metrics-block) below for more details.
* `s3_bucket_destination` - (Optional) Bucket where the S3 Storage Lens metrics export will be located. See [`s3_bucket_destination`](#s3_bucket_destination-block) below for more details.
* `storage_lens_table_destination` - (Optional) S3 table bucket where the S3 Storage Lens metrics export will be located. See [`storage_lens_table_destination`](#storage_lens_table_destination-block) below for more details.

### `expanded_prefixes_data_export` Block

The `expanded_prefixes_data_export` block supports the following:

* `s3_bucket_destination` - (Optional) Bucket where the S3 Storage Lens expanded prefix metrics export will be located. See [`s3_bucket_destination`](#s3_bucket_destination-block) below for more details.
* `storage_lens_table_destination` - (Optional) S3 table bucket where the S3 Storage Lens expanded prefix metrics export will be located. See [`storage_lens_table_destination`](#storage_lens_table_destination-block) below for more details.

### `cloud_watch_metrics` Block

The `cloud_watch_metrics` block supports the following:

* `enabled` - (Required) Whether CloudWatch publishing for S3 Storage Lens metrics is enabled.

### `s3_bucket_destination` Block

The `s3_bucket_destination` block supports the following:

* `account_id` - (Required) Account ID of the owner of the S3 Storage Lens metrics export bucket.
* `arn` - (Required) Amazon Resource Name (ARN) of the bucket.
* `encryption` - (Optional) Encryption of the metrics exports in this bucket. See [`encryption`](#encryption-block) below for more details.
* `format` - (Required) Export format. Valid values: `CSV`, `Parquet`.
* `output_schema_version` - (Required) Schema version of the export file. Valid values: `V_1`.
* `prefix` - (Optional) Prefix of the destination bucket where the metrics export will be delivered.

### `storage_lens_table_destination` Block

The `storage_lens_table_destination` block supports the following:

* `enabled` - (Required) Whether S3 Storage Lens export to S3 tables is enabled.
* `encryption` - (Optional) Encryption of the metrics exports in this S3 tables bucket. See [`encryption`](#encryption-block) below for more details.

### `encryption` Block

The `encryption` block supports the following:

* `sse_kms` - (Optional) SSE-KMS encryption. See [`sse_kms`](#sse_kms-block) below for more details.
* `sse_s3` - (Optional) SSE-S3 encryption. An empty configuration block `{}` should be used.

### `sse_kms` Block

The `sse_kms` block supports the following:

* `key_id` - (Required) KMS key ARN.

### `exclude` Block

The `exclude` block supports the following:

* `buckets` - (Optional) List of S3 bucket ARNs.
* `regions` - (Optional) List of AWS Regions.

### `include` Block

The `include` block supports the following:

* `buckets` - (Optional) List of S3 bucket ARNs.
* `regions` - (Optional) List of AWS Regions.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the S3 Storage Lens configuration.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 Storage Lens configurations using the `account_id` and `config_id`, separated by a colon (`:`). For example:

```terraform
import {
  to = aws_s3control_storage_lens_configuration.example
  id = "123456789012:example-1"
}
```

Using `terraform import`, import S3 Storage Lens configurations using the `account_id` and `config_id`, separated by a colon (`:`). For example:

```console
% terraform import aws_s3control_storage_lens_configuration.example 123456789012:example-1
```
