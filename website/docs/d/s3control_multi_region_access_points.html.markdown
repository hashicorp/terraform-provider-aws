---
subcategory: "S3 Control"
layout: "aws"
page_title: "AWS: aws_s3control_multi_region_access_points"
description: |-
  Provides details about AWS S3 Control Multi-Region Access Points.
---

# Data Source: aws_s3control_multi_region_access_points

Provides details about AWS S3 Control Multi-Region Access Points.

## Example Usage

### Basic Usage

```terraform
data "aws_s3control_multi_region_access_points" "example" {}
```

## Argument Reference

The following arguments are optional:

* `account_id` - (Optional) AWS account ID for the account that owns the multi-region access points. If omitted, defaults to the caller's account ID.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `access_points` - List of multi-region access points. See [`access_points`](#access_points-attribute-reference) below.

### `access_points` Attribute Reference

* `alias` - Alias for the multi-region access point.
* `created_at` - Time the multi-region access point was created.
* `name` - Name of the multi-region access point.
* `public_access_block` - Public access block configuration for this multi-region access point. See [`public_access_block`](#public_access_block-attribute-reference) below.
* `regions` - List of AWS Regions where the multi-region access point has data support. See [`regions`](#regions-attribute-reference) below.
* `status` - Current status of the multi-region access point.

#### `public_access_block` Attribute Reference

* `block_public_acls` - Whether Amazon S3 should block public ACLs for buckets in this account.
* `block_public_policy` - Whether Amazon S3 should block public bucket policies for buckets in this account.
* `ignore_public_acls` - Whether Amazon S3 should ignore public ACLs for buckets in this account.
* `restrict_public_buckets` - Whether Amazon S3 should restrict public bucket policies for buckets in this account.

#### `regions` Attribute Reference

* `bucket` - Name of the associated bucket for the Region.
* `bucket_account_id` - AWS account ID that owns the Amazon S3 bucket associated with this multi-region access point.
* `region` - Name of the Region.
