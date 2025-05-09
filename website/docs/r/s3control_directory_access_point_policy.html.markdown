---
subcategory: "S3 Control"
layout: "aws"
page_title: "AWS: aws_s3control_directory_access_point_policy"
description: |-
    Provides a resource to manage an access point policy for an access point resource.
---

# Resource: aws_s3control_directory_access_point_policy

Provides a resource to manage an access point policy for an access point resource.

-> For all the services in AWS Local Zones, including Amazon S3, your accountID must be enabled before you can create or access any resource in the Local Zone. You can use the `DescribeAvailabilityZones` API operation to confirm your accountID access to a Local Zone. For more information, see [AWS Documentation](https://docs.aws.amazon.com/AmazonS3/latest/userguide/opt-in-directory-bucket-lz.html).

-> In access points for directory buckets, the access point policy controls only the principal and access point resource. To control access to prefixes or APIs, you must use [`aws_s3control_directory_access_point_scope`](aws_s3control_directory_access_point_scope.html).

-> Terraform provides two ways to manage access point policies. You can use a standalone resource `aws_s3control_directory_access_point_policy` or, an in-line policy with the [`aws_s3_directory_access_point`](aws_s3_directory_access_point.html) resource. You cannot use a standalone resource at the same time as in-line, which will cause an overwrite of each other. You must use one or the other.

-> Bucket type: This resource cannot be used for access point policies for general purpose buckets, see [`aws_s3control_access_point_policy`](aws_s3control_access_point_policy.html) for more. 

## Example Usage

### S3 Access Point Policy for a directory bucket in an AWS Local Zone

```terraform
resource "aws_s3_directory_bucket" "example" {
  bucket = "example--zoneid--x-s3"
}


resource "aws_s3_directory_access_point" "example" {
  bucket = aws_s3_directory_bucket.example.id
  name = "example--zoneid--xa-s3"
  account_id = "123456789012"
}

resource "aws_s3control_directory_access_point_policy" "example" {
  access_point_arn = aws_s3_directory_access_point.example.arn

  policy = jsonencode({
    Version = "2008-10-17"
    Statement = [{
      Effect = "Allow"
      Action = "s3express:CreateSession"
      Principal = {
        AWS = "arn:aws:iam::123456789012:root"
      }
      Resource = "${aws_s3_directory_access_point.example.arn}"
    }]
  })
}
```

## Argument Reference

This resource supports the following arguments:

* `access_point_arn` - (Required) The access point ARN that you want to associate with the specified policy.
* `policy` - (Required) The policy that you want to apply to the specified access point.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:
* `id` - The access point name and AWS account ID separated by a colon (`:`).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Access Point policies using the `access_point_arn`. For example:

```terraform
import {
  to = aws_s3control_directory_access_point_policy.example
  id = "example--zoneid--xa-s3:123456789012"
}
```

Using `terraform import`, import Access Point policies using the `access_point_arn`. For example:

```console
% terraform import aws_s3control_directory_access_point_policy.example example--zoneid--xa-s3:123456789012
```