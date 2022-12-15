---
subcategory: "S3 Control"
layout: "aws"
page_title: "AWS: aws_s3control_multi_region_access_point"
description: |-
  Provides details on S3 multi region access point
---

# Data Source: aws_s3control_multi_region_access_point

Provides details on a specific S3 multi region access point

## Example Usage

The following example retrieves IAM policy of a specified S3 bucket.

```terraform
data "aws_s3control_multi_region_access_point" "example" {
  name = "example"
}
```

## Argument Reference

The following arguments are supported:

* `account_id` - (Optional) The AWS account ID of the S3 multi region access point. Defaults to automatically determined account ID of the Terraform AWS provider.
* `name` - (Required) The name of the Multi-Region Access Point.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `created_at` - Timestamp when the resource has been created.
* `name` - Name of the S3 multi region access point.
* `public_access_block` - Public Access Block of the S3 multi region access point. Detailed below.
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

* `Bucket` - The name of the bucket.
* `Region` - The name of the region.
