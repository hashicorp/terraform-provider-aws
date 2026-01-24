---
subcategory: "S3 Tables"
layout: "aws"
page_title: "AWS: aws_s3tables_table_bucket"
description: |-
  Terraform resource for managing an Amazon S3 Tables Table Bucket.
---

# Resource: aws_s3tables_table_bucket

Terraform resource for managing an Amazon S3 Tables Table Bucket.

## Example Usage

### Basic Usage

```terraform
resource "aws_s3tables_table_bucket" "example" {
  name = "example-bucket"
}
```

### With Storage Class Configuration

```terraform
resource "aws_s3tables_table_bucket" "example" {
  name = "example-bucket"

  storage_class_configuration = {
    storage_class = "INTELLIGENT_TIERING"
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required, Forces new resource) Name of the table bucket.
  Must be between 3 and 63 characters in length.
  Can consist of lowercase letters, numbers, and hyphens, and must begin and end with a lowercase letter or number.
  A full list of bucket naming rules can be found in the [S3 Tables documentation](https://docs.aws.amazon.com/AmazonS3/latest/userguide/s3-tables-buckets-naming.html#table-buckets-naming-rules).

The following arguments are optional:

* `encryption_configuration` - (Optional) A single table bucket encryption configuration object.
  [See `encryption_configuration` below](#encryption_configuration).
* `force_destroy` - (Optional, Default:`false`) Whether all tables and namespaces within the table bucket should be deleted *when the table bucket is destroyed* so that the table bucket can be destroyed without error. These tables and namespaces are *not* recoverable. This only deletes tables and namespaces when the table bucket is destroyed, *not* when setting this parameter to `true`. Once this parameter is set to `true`, there must be a successful `terraform apply` run before a destroy is required to update this value in the resource state. Without a successful `terraform apply` after this parameter is set, this flag will have no effect. If setting this field in the same operation that would require replacing the table bucket or destroying the table bucket, this flag will not work. Additionally when importing a table bucket, a successful `terraform apply` is required to set this value in state before it will take effect on a destroy operation.
* `maintenance_configuration` - (Optional) A single table bucket maintenance configuration object.
  [See `maintenance_configuration` below](#maintenance_configuration).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `storage_class_configuration` - (Optional) A single table bucket storage class configuration object.
  If not specified, AWS will use the default storage class (`STANDARD`).
  [See `storage_class_configuration` below](#storage_class_configuration).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `encryption_configuration`

The `encryption_configuration` object supports the following arguments:

* `kms_key_arn` - (Optional) The ARN of a KMS Key to be used with `aws:kms` `sse_algorithm`
* `sse_algorithm` - (Required) One of `aws:kms` or `AES256`

### `maintenance_configuration`

The `maintenance_configuration` object supports the following argument:

* `iceberg_unreferenced_file_removal` - (Required) A single Iceberg unreferenced file removal settings object.
  [See `iceberg_unreferenced_file_removal` below](#iceberg_unreferenced_file_removal).

### `iceberg_unreferenced_file_removal`

The `iceberg_unreferenced_file_removal` object supports the following arguments:

* `settings` - (Required) Settings object for unreferenced file removal.
  [See `iceberg_unreferenced_file_removal.settings` below](#iceberg_unreferenced_file_removalsettings).
* `status` - (Required) Whether the configuration is enabled.
  Valid values are `enabled` and `disabled`.

### `iceberg_unreferenced_file_removal.settings`

The `iceberg_unreferenced_file_removal.settings` object supports the following arguments:

* `non_current_days` - (Required) Data objects marked for deletion are deleted after this many days.
  Must be at least `1`.
* `unreferenced_days` - (Required) Unreferenced data objects are marked for deletion after this many days.
  Must be at least `1`.

### `storage_class_configuration`

The `storage_class_configuration` object supports the following argument:

* `storage_class` - (Required) The storage class for the table bucket. Valid values are `STANDARD` and `INTELLIGENT_TIERING`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the table bucket.
* `created_at` - Date and time when the bucket was created.
* `owner_account_id` - Account ID of the account that owns the table bucket.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 Tables Table Bucket using the `arn`. For example:

```terraform
import {
  to = aws_s3tables_table_bucket.example
  id = "arn:aws:s3tables:us-west-2:123456789012:bucket/example-bucket"
}
```

Using `terraform import`, import S3 Tables Table Bucket using the `arn`. For example:

```console
% terraform import aws_s3tables_table_bucket.example arn:aws:s3tables:us-west-2:123456789012:bucket/example-bucket
```
