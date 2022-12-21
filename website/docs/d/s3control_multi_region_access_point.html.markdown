---
subcategory: "S3 Control"
layout: "aws"
page_title: "AWS: aws_s3control_multi_region_access_point"
description: |-
  Provides details an S3 Multi-Region Access Point.
---

# Data Source: aws_s3control_multi_region_access_point

Provides details on a specific S3 Multi-Region Access Point.

## Example Usage

```terraform
data "aws_s3control_multi_region_access_point" "example" {
  name = "example"
}
```

## Argument Reference

The following arguments are supported:

* `account_id` - (Optional) The AWS account ID of the S3 Multi-Region Access Point. Defaults to automatically determined account ID of the Terraform AWS provider.
* `name` - (Required) The name of the Multi-Region Access Point.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `alias` - The alias for the Multi-Region Access Point.
* `arn` - Amazon Resource Name (ARN) of the Multi-Region Access Point.
* `created_at` - Timestamp when the resource has been created.
* `domain_name` - The DNS domain name of the S3 Multi-Region Access Point in the format _`alias`_.accesspoint.s3-global.amazonaws.com. For more information, see the documentation on [Multi-Region Access Point Requests](https://docs.aws.amazon.com/AmazonS3/latest/userguide/MultiRegionAccessPointRequests.html).
* `public_access_block` - Public Access Block of the Multi-Region Access Point. Detailed below.
* `regions` - A collection of the regions and buckets associated with the Multi-Region Access Point.
* `status` - The current status of the Multi-Region Access Point.

### public_access_block

* `block_public_acls` - Specifies whether Amazon S3 should block public access control lists (ACLs). When set to `true` causes the following behavior:
    * PUT Bucket acl and PUT Object acl calls fail if the specified ACL is public.
    * PUT Object calls fail if the request includes a public ACL.
    * PUT Bucket calls fail if the request includes a public ACL.
* `block_public_policy` - Specifies whether Amazon S3 should block public bucket policies for buckets in this account. When set to `true` causes Amazon S3 to:
    * Reject calls to PUT Bucket policy if the specified bucket policy allows public access.
* `ignore_public_acls` - Specifies whether Amazon S3 should ignore public ACLs for buckets in this account. When set to `true` causes Amazon S3 to:
    * Ignore all public ACLs on buckets in this account and any objects that they contain.
* `restrict_public_buckets` - Specifies whether Amazon S3 should restrict public bucket policies for buckets in this account. When set to `true`:
    * Only the bucket owner and AWS Services can access buckets with public policies.

### regions

* `bucket` - The name of the bucket.
* `region` - The name of the region.
