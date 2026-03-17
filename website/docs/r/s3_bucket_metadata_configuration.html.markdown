---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_bucket_metadata_configuration"
description: |-
  Manages Amazon S3 Metadata for a bucket.
---

# Resource: aws_s3_bucket_metadata_configuration

Manages Amazon S3 Metadata for a bucket.

## Example Usage

### Basic Usage

```terraform
resource "aws_s3_bucket_metadata_configuration" "example" {
  bucket = aws_s3_bucket.example.bucket

  metadata_configuration {
    inventory_table_configuration {
      configuration_state = "ENABLED"
    }

    journal_table_configuration {
      record_expiration {
        days       = 7
        expiration = "ENABLED"
      }
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `bucket` - (Required) General purpose bucket that you want to create the metadata configuration for.
* `metadata_configuration` - (Required) Metadata configuration. See [`metadata_configuration` Block](#metadata_configuration-block) for details.

The following arguments are optional:

* `expected_bucket_owner` - (Optional, Forces new resource, **Deprecated**) Account ID of the expected bucket owner.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `metadata_configuration` Block

The `metadata_configuration` configuration block supports the following arguments:

* `inventory_table_configuration` - (Required) Inventory table configuration. See [`inventory_table_configuration` Block](#inventory_table_configuration-block) for details.
* `journal_table_configuration` - (Required) Journal table configuration. See [`journal_table_configuration` Block](#journal_table_configuration-block) for details.

### `inventory_table_configuration` Block

The `inventory_table_configuration` configuration block supports the following arguments:

* `configuration_state` - (Required) Configuration state of the inventory table, indicating whether the inventory table is enabled or disabled. Valid values: `ENABLED`, `DISABLED`.
* `encryption_configuration` - (Optional) Encryption configuration for the inventory table. See [`encryption_configuration` Block](#encryption_configuration-block) for details.

### `journal_table_configuration` Block

The `journal_table_configuration` configuration block supports the following arguments:

* `encryption_configuration` - (Optional) Encryption configuration for the journal table. See [`encryption_configuration` Block](#encryption_configuration-block) for details.
* `record_expiration` - (Required) Journal table record expiration settings. See [`record_expiration` Block](#record_expiration-block) for details.

### `encryption_configuration` Block

The `encryption_configuration` configuration block supports the following arguments:

* `kms_key_arn` - (Optional) KMS key ARN when `sse_algorithm` is `aws:kms`.
* `sse_algorithm` - (Required) Encryption type for the metadata table. Valid values: `aws:kms`, `AES256`.

### `record_expiration` Block

The `record_expiration` configuration block supports the following arguments:

* `days` - (Optional) Number of days to retain journal table records.
* `expiration` - (Required) Whether journal table record expiration is enabled or disabled. Valid values: `ENABLED`, `DISABLED`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `metadata_configuration.0.destination` - Destination information for the S3 Metadata configuration.
    * `table_bucket_arn` - ARN of the table bucket where the metadata configuration is stored.
    * `table_bucket_type` - Type of the table bucket where the metadata configuration is stored.
    * `table_namespace` - Namespace in the table bucket where the metadata tables for the metadata configuration are stored.
* `metadata_configuration.0.inventory_table_configuration.0.table_arn` - Inventory table ARN.
* `metadata_configuration.0.inventory_table_configuration.0.table_name` - Inventory table name.
* `metadata_configuration.0.journal_table_configuration.0.table_arn` - Journal table ARN.
* `metadata_configuration.0.journal_table_configuration.0.table_name` - Journal table name.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_s3_bucket_metadata_configuration.example
  identity = {
    bucket = "bucket-name"
  }
}

resource "aws_s3_bucket_metadata_configuration" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `bucket` (String) S3 bucket name.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 bucket metadata configuration using the `bucket`. For example:

```terraform
import {
  to = aws_s3_bucket_metadata_configuration.example
  id = "bucket-name"
}
```

**Using `terraform import` to import** S3 bucket metadata configuration using the `bucket`. For example:

```console
% terraform import aws_s3_bucket_metadata_configuration.example bucket-name
```
