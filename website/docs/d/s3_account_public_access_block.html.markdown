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

output "foo" {
  block_public_acls_value = data.aws_s3_bucket_policy.example.block_public_acls
  block_public_policy_value = data.aws_s3_bucket_policy.example.block_public_policy
  ignore_public_acls_value = data.aws_s3_bucket_policy.example.ignore_public_acls
  restrict_public_buckets_value = data.aws_s3_bucket_policy.example.restrict_public_buckets
}
```

## Argument Reference

The following arguments are supported:

* `account_id` - (Optional) AWS account ID to configure. Defaults to automatically determined account ID of the Terraform AWS provider.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - AWS account ID
* `block_public_acls` - Whether or not Amazon S3 should block public ACLs for buckets in this account is enabled. Returns as `true` or `false`.
* `block_public_policy` - Whether or not Amazon S3 should block public bucket policies for buckets in this account is enabled. Returns as `true` or `false`.
* `ignore_public_acls` - Whether or not Amazon S3 should ignore public ACLs for buckets in this account is enabled. Returns as `true` or `false`.
* `restrict_public_buckets` - Whether or not Amazon S3 should restrict public bucket policies for buckets in this account is enabled. Returns as `true` or `false`.
