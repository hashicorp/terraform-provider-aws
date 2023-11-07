---
subcategory: "S3 Control"
layout: "aws"
page_title: "AWS: aws_s3_account_public_access_block"
description: |-
  Provides S3 account-level Public Access Block Configuration
---

# Data Source: aws_s3_account_public_access_block

The S3 account public access block data source returns account-level public access block configuration.

## Example Usage

```terraform
data "aws_s3_account_public_access_block" "example" {
}
```

## Argument Reference

This data source supports the following arguments:

* `account_id` - (Optional) AWS account ID to configure. Defaults to automatically determined account ID of the Terraform AWS provider.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - AWS account ID
* `block_public_acls` - Whether or not Amazon S3 should block public ACLs for buckets in this account is enabled. Returns as `true` or `false`.
* `block_public_policy` - Whether or not Amazon S3 should block public bucket policies for buckets in this account is enabled. Returns as `true` or `false`.
* `ignore_public_acls` - Whether or not Amazon S3 should ignore public ACLs for buckets in this account is enabled. Returns as `true` or `false`.
* `restrict_public_buckets` - Whether or not Amazon S3 should restrict public bucket policies for buckets in this account is enabled. Returns as `true` or `false`.
