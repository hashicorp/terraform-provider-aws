---
subcategory: "S3 Control"
layout: "aws"
page_title: "AWS: aws_s3control_directory_bucket_access_point_scope"
description: |-
  Provides a resource to manage the access point scope for a directory bucket.
---

# Resource: aws_s3control_directory_bucket_access_point_scope

Provides a resource to manage the access point scope for a directory bucket.

With access points for directory buckets, you can use the access point scope to restrict access to specific prefixes, API actions, or a combination of both. You can specify any amount of prefixes, but the total length of characters of all prefixes must be less than 256 bytes. For more information, see [AWS Documentation](https://docs.aws.amazon.com/AmazonS3/latest/userguide/access-points-directory-buckets-manage-scope.html).

-> For all the services in AWS Local Zones, including Amazon S3, your accountID must be enabled before you can create or access any resource in the Local Zone. You can use the `DescribeAvailabilityZones` API operation to confirm your accountID access to a Local Zone. For more information, see [AWS Documentation](https://docs.aws.amazon.com/AmazonS3/latest/userguide/opt-in-directory-bucket-lz.html)

-> Terraform provides two ways to manage access point scopes. You can use a standalone resource `aws_s3control_directory_access_point_scope` or, an in-line scope with the  [`aws_s3_directory_access_point`](aws_s3_directory_access_point.html) resource. You cannot use a standalone resource at the same time as in-line, which will cause an overwrite of each other. You must use one or the other.

## Example Usage

### S3 Access Point Scope for a directory bucket in an AWS Local Zone

```terraform
data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_s3_directory_bucket" "example" {
  bucket = "example--zoneId--x-s3"
  location {
    name = data.aws_availability_zones.available.zone_ids[0]
  }
}

resource "aws_s3_access_point" "example" {
  bucket = aws_s3_directory_bucket.example.id
  name   = "example--zoneId--xa-s3"
}

resource "aws_s3control_directory_bucket_access_point_scope" "example" {
  name       = "example--zoneId--xa-s3"
  account_id = "123456789012"
  scope {
    permissions = ["GetObject", "ListBucket"]
    prefixes    = ["myobject1.csv", "myobject2*"]
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `account_id` - (Required) The AWS account ID that owns the specified access point.
* `name` - (Required) The name of the access point that you want to apply the scope to.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `scope` - (Optional). Scope is used to restrict access to specific prefixes, API operations, or a combination of both. To remove the `scope`, set it to `{permissions=[] prefixes=[]}`. The default scope is `{permissions=[] prefixes=[]}`.

### Scope Configuration block

The following arguments are optional:

* `permissions` – (Optional) You can specify a list of API operations as permissions for the access point.
* `prefixes` – (Optional) You can specify a list of prefixes, but the total length of characters of all prefixes must be less than 256 bytes.

* For more information on access point scope, see [AWS Documentation](https://docs.aws.amazon.com/AmazonS3/latest/userguide/access-points-directory-buckets-manage-scope.html).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import access point scope using the access point name and AWS accountID separated by a colon (`:`). For example:

```terraform
import {
  to = aws_s3control_directory_bucket_access_point_scope.example
  id = "example--zoneid--xa-s3,123456789012"
}
```

Using `terraform import`, import Access Point Scope using access point name and AWS account ID separated by a colon (`,`). For example:

```console
% terraform import aws_s3control_directory_bucket_access_point_scope.example example--zoneid--xa-s3,123456789012
```
