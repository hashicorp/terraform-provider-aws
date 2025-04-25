---
subcategory: "S3 Control"
layout: "aws"
page_title: "AWS: aws_s3control_directory_access_point_policy"
description: |-
    Provides a resource to manage access point policy for an S3 Access Point for an S3 directory bucket resource.
---

# Resource: aws_s3control_directory_access_point_policy

Provides a resource to manage access point policy for an S3 Access Point for an S3 directory bucket resource.

NOTE: In access points for directory buckets, the access point policy controls only the principal and access point resource. To control access to prefixes or APIs, you must use [`aws_s3control_directory_access_point_scope`](aws_s3control_directory_access_point_scope.html).

NOTE: Terraform provides two ways to manage S3 Access Points and their policies. You can use a standalone resource (`aws_s3control_directory_access_point_policy`) or, you can use an in-line policy with the access point for directory bucket resource [`aws_s3_directory_access_point`](aws_s3_directory_access_point.html). You cannot use a standalone resource at the same time as in-line, which will cause an overwrite of each other. You must use one or the other.

## Example Usage

### S3 Access Point Policy for a directory bucket in an AWS Availability Zone

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