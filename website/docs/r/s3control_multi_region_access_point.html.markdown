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

For more information, see the documentation on [Multi-Region Access Points](https://docs.aws.amazon.com/AmazonS3/latest/userguide/MultiRegionAccessPoints.html).

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

* `account_id` - (Optional) AWS account ID for the owner of the buckets for which you want to create a Multi-Region Access Point. Defaults to automatically determined account ID of the Terraform AWS provider.
* `details` - (Required) Configuration block containing details about the Multi-Region Access Point. See [`details` Block](#details) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `details` Block

The `details` configuration block supports the following arguments:

* `name` - (Required) Name of the Multi-Region Access Point.
* `public_access_block` - (Optional) Configuration block to manage the `PublicAccessBlock` configuration that you want to apply to this Multi-Region Access Point. You can enable the configuration options in any combination. See [`public_access_block` Block](#public_access_block) below.
* `region` - (Required) Region configuration block to specify the bucket associated with the Multi-Region Access Point. See [`region` Block](#region) below.

### `public_access_block` Block

The `public_access_block` configuration block supports the following arguments:

* `block_public_acls` - (Optional) Whether Amazon S3 should block public ACLs for buckets in this account. Defaults to `true`. Enabling this setting does not affect existing policies or ACLs. When set to `true`, PUT Bucket acl and PUT Object acl calls fail if the specified ACL is public, PUT Object calls fail if the request includes a public ACL, and PUT Bucket calls fail if the request includes a public ACL.
* `block_public_policy` - (Optional) Whether Amazon S3 should block public bucket policies for buckets in this account. Defaults to `true`. Enabling this setting does not affect existing bucket policies. When set to `true`, Amazon S3 rejects calls to PUT Bucket policy if the specified bucket policy allows public access.
* `ignore_public_acls` - (Optional) Whether Amazon S3 should ignore public ACLs for buckets in this account. Defaults to `true`. Enabling this setting does not affect the persistence of any existing ACLs and doesn't prevent new public ACLs from being set. When set to `true`, Amazon S3 ignores all public ACLs on buckets in this account and any objects that they contain.
* `restrict_public_buckets` - (Optional) Whether Amazon S3 should restrict public bucket policies for buckets in this account. Defaults to `true`. Enabling this setting does not affect previously stored bucket policies, except that public and cross-account access within any public bucket policy, including non-public delegation to specific accounts, is blocked. When set to `true`, only the bucket owner and AWS Services can access buckets with public policies.

### `region` Block

The `region` configuration block supports the following arguments:

* `bucket` - (Required) Name of the associated bucket for the Region.
* `bucket_account_id` - (Optional) AWS account ID that owns the Amazon S3 bucket that's associated with this Multi-Region Access Point.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `alias` - Alias for the Multi-Region Access Point.
* `arn` - Amazon Resource Name (ARN) of the Multi-Region Access Point.
* `domain_name` - DNS domain name of the S3 Multi-Region Access Point in the format _`alias`_.accesspoint.s3-global.amazonaws.com. For more information, see the documentation on [Multi-Region Access Point Requests](https://docs.aws.amazon.com/AmazonS3/latest/userguide/MultiRegionAccessPointRequests.html).
* `id` - AWS account ID and access point name separated by a colon (`:`).
* `name` - Name of the Multi-Region Access Point.
* `status` - Status of the Multi-Region Access Point. One of: `READY`, `INCONSISTENT_ACROSS_REGIONS`, `CREATING`, `PARTIALLY_CREATED`, `PARTIALLY_DELETED`, `DELETING`.

### `region` Block

* `region` - Name of the Region.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `delete` - (Default `15m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_s3control_multi_region_access_point.example
  identity = {
    name = "example"
  }
}

resource "aws_s3control_multi_region_access_point" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `name` (String) Name of the Multi-Region Access Point.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

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
