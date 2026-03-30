---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_bucket_abac"
description: |-
  Manages ABAC (Attribute Based Access Control) for an AWS S3 (Simple Storage) Bucket.
---

# Resource: aws_s3_bucket_abac

Manages ABAC (Attribute Based Access Control) for an AWS S3 (Simple Storage) Bucket.
See the [AWS documentation](https://docs.aws.amazon.com/AmazonS3/latest/userguide/buckets-tagging-enable-abac.html) on enabling ABAC for general purpose buckets for additional information.

## Example Usage

### Basic Usage

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "bucket-name"
}

resource "aws_s3_bucket_abac" "example" {
  bucket = aws_s3_bucket.example.bucket

  abac_status {
    status = "Enabled"
  }
}
```

## Argument Reference

The following arguments are required:

* `bucket` - (Required) General purpose bucket that you want to create the metadata configuration for.
* `abac_status` - (Required) ABAC status configuration. See [`abac_status` Block](#abac_status-block) for details.

The following arguments are optional:

* `expected_bucket_owner` - (Optional, Forces new resource, **Deprecated**) Account ID of the expected bucket owner.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `abac_status` Block

The `abac_status` configuration block supports the following arguments:

* `status` - (Required) ABAC status of the general purpose bucket.
Valid values are `Enabled` and `Disabled`.
By default, ABAC is disabled for all Amazon S3 general purpose buckets.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 (Simple Storage) Bucket ABAC using the `bucket` or `bucket` and `expected_bucket_owner` separated by a comma (`,`). For example:

If the owner (account ID) of the source bucket is the same account used to configure the Terraform AWS Provider, import using the `bucket`:

```terraform
import {
  to = aws_s3_bucket_abac.example
  id = "bucket-name"
}
```

If the owner (account ID) of the source bucket differs from the account used to configure the Terraform AWS Provider, import using the `bucket` and `expected_bucket_owner` separated by a comma (`,`):

```terraform
import {
  to = aws_s3_bucket_abac.example
  id = "bucket-name,123456789012"
}
```

Using `terraform import`, import S3 (Simple Storage) Bucket ABAC using the `bucket` or `bucket` and `expected_bucket_owner` separated by a comma (`,`). For example:

If the owner (account ID) of the source bucket is the same account used to configure the Terraform AWS Provider, import using the `bucket`:

```console
% terraform import aws_s3_bucket_abac.example bucket-name
```

If the owner (account ID) of the source bucket differs from the account used to configure the Terraform AWS Provider, import using the `bucket` and `expected_bucket_owner` separated by a comma (`,`):

```console
% terraform import aws_s3_bucket_abac.example bucket-name,123456789012
```
