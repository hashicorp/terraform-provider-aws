---
subcategory: "S3 Control"
layout: "aws"
page_title: "AWS: aws_s3control_multi_region_access_point"
description: |-
  Provides a resource to manage an S3 Multi-Region Access Point associated with specified buckets.
---

# Resource: aws_s3control_multi_region_access_point

Provides a resource to manage an S3 Multi-Region Access Point associated with specified buckets.

-> This resource cannot be used with S3 directory buckets.

## Example Usage

### Multiple AWS Buckets in Different Regions

```terraform
provider "aws" {
  region = "us-east-1"
  alias  = "primary_region"
}

provider "aws" {
  region = "us-west-2"
  alias  = "secondary_region"
}

resource "aws_s3_bucket" "foo_bucket" {
  provider = aws.primary_region

  bucket = "example-bucket-foo"
}

resource "aws_s3_bucket" "bar_bucket" {
  provider = aws.secondary_region

  bucket = "example-bucket-bar"
}

resource "aws_s3control_multi_region_access_point" "example" {
  details {
    name = "example"

    region {
      bucket = aws_s3_bucket.foo_bucket.id
    }

    region {
      bucket = aws_s3_bucket.bar_bucket.id
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `account_id` - (Optional) The AWS account ID for the owner of the buckets for which you want to create a Multi-Region Access Point. Defaults to automatically determined account ID of the Terraform AWS provider.
* `details` - (Required) A configuration block containing details about the Multi-Region Access Point. See [Details Configuration Block](#details-configuration) below for more details

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `delete` - (Default `15m`)

### Details Configuration

The `details` block supports the following:

* `name` - (Required) The name of the Multi-Region Access Point.
* `public_access_block` - (Optional) Configuration block to manage the `PublicAccessBlock` configuration that you want to apply to this Multi-Region Access Point. You can enable the configuration options in any combination. See [Public Access Block Configuration](#public-access-block-configuration) below for more details.
* `region` - (Required) The Region configuration block to specify the bucket associated with the Multi-Region Access Point. See [Region Configuration](#region-configuration) below for more details.

For more information, see the documentation on [Multi-Region Access Points](https://docs.aws.amazon.com/AmazonS3/latest/userguide/MultiRegionAccessPoints.html).

### Public Access Block Configuration

The `public_access_block` block supports the following:

* `block_public_acls` - (Optional) Whether Amazon S3 should block public ACLs for buckets in this account. Defaults to `true`. Enabling this setting does not affect existing policies or ACLs. When set to `true` causes the following behavior:
    * PUT Bucket acl and PUT Object acl calls fail if the specified ACL is public.
    * PUT Object calls fail if the request includes a public ACL.
    * PUT Bucket calls fail if the request includes a public ACL.
* `block_public_policy` - (Optional) Whether Amazon S3 should block public bucket policies for buckets in this account. Defaults to `true`. Enabling this setting does not affect existing bucket policies. When set to `true` causes Amazon S3 to:
    * Reject calls to PUT Bucket policy if the specified bucket policy allows public access.
* `ignore_public_acls` - (Optional) Whether Amazon S3 should ignore public ACLs for buckets in this account. Defaults to `true`. Enabling this setting does not affect the persistence of any existing ACLs and doesn't prevent new public ACLs from being set. When set to `true` causes Amazon S3 to:
    * Ignore all public ACLs on buckets in this account and any objects that they contain.
* `restrict_public_buckets` - (Optional) Whether Amazon S3 should restrict public bucket policies for buckets in this account. Defaults to `true`. Enabling this setting does not affect previously stored bucket policies, except that public and cross-account access within any public bucket policy, including non-public delegation to specific accounts, is blocked. When set to `true`:
    * Only the bucket owner and AWS Services can access buckets with public policies.

### Region Configuration

The `region` block supports the following:

* `bucket` - (Required) The name of the associated bucket for the Region.
* `bucket_account_id` - (Optional) The AWS account ID that owns the Amazon S3 bucket that's associated with this Multi-Region Access Point.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `alias` - The alias for the Multi-Region Access Point.
* `arn` - Amazon Resource Name (ARN) of the Multi-Region Access Point.
* `domain_name` - The DNS domain name of the S3 Multi-Region Access Point in the format _`alias`_.accesspoint.s3-global.amazonaws.com. For more information, see the documentation on [Multi-Region Access Point Requests](https://docs.aws.amazon.com/AmazonS3/latest/userguide/MultiRegionAccessPointRequests.html).
* `id` - The AWS account ID and access point name separated by a colon (`:`).
* `status` - The current status of the Multi-Region Access Point. One of: `READY`, `INCONSISTENT_ACROSS_REGIONS`, `CREATING`, `PARTIALLY_CREATED`, `PARTIALLY_DELETED`, `DELETING`.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Multi-Region Access Points using the `account_id` and `name` of the Multi-Region Access Point separated by a colon (`:`). For example:

```terraform
import {
  to = aws_s3control_multi_region_access_point.example
  id = "123456789012:example"
}
```

Using `terraform import`, import Multi-Region Access Points using the `account_id` and `name` of the Multi-Region Access Point separated by a colon (`:`). For example:

```console
% terraform import aws_s3control_multi_region_access_point.example 123456789012:example
```
