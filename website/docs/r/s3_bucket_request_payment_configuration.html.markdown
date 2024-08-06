---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_bucket_request_payment_configuration"
description: |-
  Provides an S3 bucket request payment configuration resource.
---

# Resource: aws_s3_bucket_request_payment_configuration

Provides an S3 bucket request payment configuration resource. For more information, see [Requester Pays Buckets](https://docs.aws.amazon.com/AmazonS3/latest/dev/RequesterPaysBuckets.html).

~> **NOTE:** Destroying an `aws_s3_bucket_request_payment_configuration` resource resets the bucket's `payer` to the S3 default: the bucket owner.

-> This resource cannot be used with S3 directory buckets.

## Example Usage

```terraform
resource "aws_s3_bucket_request_payment_configuration" "example" {
  bucket = aws_s3_bucket.example.id
  payer  = "Requester"
}
```

## Argument Reference

This resource supports the following arguments:

* `bucket` - (Required, Forces new resource) Name of the bucket.
* `expected_bucket_owner` - (Optional, Forces new resource) Account ID of the expected bucket owner.
* `payer` - (Required) Specifies who pays for the download and request fees. Valid values: `BucketOwner`, `Requester`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The `bucket` or `bucket` and `expected_bucket_owner` separated by a comma (`,`) if the latter is provided.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 bucket request payment configuration using the `bucket` or using the `bucket` and `expected_bucket_owner` separated by a comma (`,`). For example:

If the owner (account ID) of the source bucket is the same account used to configure the Terraform AWS Provider, import using the `bucket`:

```terraform
import {
  to = aws_s3_bucket_request_payment_configuration.example
  id = "bucket-name"
}
```

If the owner (account ID) of the source bucket differs from the account used to configure the Terraform AWS Provider, import using the `bucket` and `expected_bucket_owner` separated by a comma (`,`):

```terraform
import {
  to = aws_s3_bucket_request_payment_configuration.example
  id = "bucket-name,123456789012"
}
```

**Using `terraform import` to import** S3 bucket request payment configuration using the `bucket` or using the `bucket` and `expected_bucket_owner` separated by a comma (`,`). For example:

If the owner (account ID) of the source bucket is the same account used to configure the Terraform AWS Provider, import using the `bucket`:

```console
% terraform import aws_s3_bucket_request_payment_configuration.example bucket-name
```

If the owner (account ID) of the source bucket differs from the account used to configure the Terraform AWS Provider, import using the `bucket` and `expected_bucket_owner` separated by a comma (`,`):

```console
% terraform import aws_s3_bucket_request_payment_configuration.example bucket-name,123456789012
```
