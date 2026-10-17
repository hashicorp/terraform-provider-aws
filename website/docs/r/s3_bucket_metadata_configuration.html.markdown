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

### Annotation Table

`annotation_table_configuration` is set as an attribute (`=`), not a block, and every argument must be present in the object literal (use `null` for any you don't want to set):

```terraform
resource "aws_iam_role" "example" {
  name = "example-s3-annotation-table-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        Service = "metadata.s3.amazonaws.com"
      }
      Action = "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_role_policy" "example" {
  role = aws_iam_role.example.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
        "s3:GetObjectAnnotation",
        "s3:GetObjectVersionAnnotation",
        "s3:ListBucket",
        "s3:ListBucketVersions",
      ]
      Resource = [aws_s3_bucket.example.arn, "${aws_s3_bucket.example.arn}/*"]
    }]
  })
}

resource "aws_s3_bucket_metadata_configuration" "example" {
  bucket = aws_s3_bucket.example.bucket

  metadata_configuration {
    annotation_table_configuration = [{
      configuration_state      = "ENABLED"
      role                     = aws_iam_role.example.arn
      encryption_configuration = null
    }]

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

* `annotation_table_configuration` - (Optional) Annotation table configuration. See [`annotation_table_configuration` Block](#annotation_table_configuration-block) for details.
* `inventory_table_configuration` - (Required) Inventory table configuration. See [`inventory_table_configuration` Block](#inventory_table_configuration-block) for details.
* `journal_table_configuration` - (Required) Journal table configuration. See [`journal_table_configuration` Block](#journal_table_configuration-block) for details.

### `annotation_table_configuration` Block

Unlike the other blocks under `metadata_configuration`, `annotation_table_configuration` is set as an attribute (`annotation_table_configuration = [{ ... }]`), not with block syntax. Every argument below must be present in the object literal; use `null` for any you don't want to set.

* `configuration_state` - (Required) Configuration state of the annotation table, indicating whether the annotation table is enabled or disabled. Valid values: `ENABLED`, `DISABLED`.
* `encryption_configuration` - (Optional) Encryption configuration for the annotation table. Set to `null` if not configuring encryption. See [`encryption_configuration` Block](#encryption_configuration-block) for details.
* `role` - (Optional) ARN of the IAM role used to manage the annotation table. Required when `configuration_state` is `ENABLED`; must be `null` when `configuration_state` is `DISABLED`. The role's trust policy must allow `metadata.s3.amazonaws.com` to assume it, and its permissions policy must grant `s3:GetObjectAnnotation`, `s3:GetObjectVersionAnnotation`, `s3:ListBucket`, and `s3:ListBucketVersions` on the bucket (plus `kms:Decrypt` if the bucket or annotations are encrypted with a KMS key).

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

### `metadata_configuration` Block

* `annotation_table_arn` - Annotation table ARN.
* `annotation_table_name` - Annotation table name.
* `destination` - Destination information for the S3 Metadata configuration. See [`destination` Block](#destination-block) for details.

### `destination` Block

* `table_bucket_arn` - ARN of the table bucket where the metadata configuration is stored.
* `table_bucket_type` - Type of the table bucket where the metadata configuration is stored.
* `table_namespace` - Namespace in the table bucket where the metadata tables for the metadata configuration are stored.

### `inventory_table_configuration` Block

* `table_arn` - Inventory table ARN.
* `table_name` - Inventory table name.

### `journal_table_configuration` Block

* `table_arn` - Journal table ARN.
* `table_name` - Journal table name.

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
