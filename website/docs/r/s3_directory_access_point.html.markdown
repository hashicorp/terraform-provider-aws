---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_directory_access_point"
description: |-
  Provides a resource to manage an access point for a directory bucket.
---
# Resource: aws_s3_directory_access_point
Provides a resource to manage an access point for a directory bucket.

-> For all the services in AWS Local Zones, including Amazon S3, your accountID must be enabled before you can create or access any resource in the Local Zone. You can use the `DescribeAvailabilityZones` API operation to confirm your accountID access to a Local Zone. For more information, see [AWS Documentation](https://docs.aws.amazon.com/AmazonS3/latest/userguide/opt-in-directory-bucket-lz.html)

-> Terraform provides two ways to manage access point policy and access point scope. You can use standalone resources [`aws_s3control_directory_access_point_policy`](aws_s3control_directory_access_point_policy.html) and  [`aws_s3control_directory_access_point_scope`](aws_s3control_directory_access_point_scope.html) or, you can use in-line resource [`aws_s3_directory_access_point`](aws_s3_directory_access_point.html). You cannot use a standalone resource at the same time as in-line, which will cause an overwrite of each other. You must use one or the other.

-> Bucket type: this resource cannot be used for access points for general purpose buckets, see [`aws_s3_access_point`](s3_access_point.html) for more. 

## Example Usage
### S3 Access Point for a directory bucket in an AWS Local Zone
```terraform
resource "aws_s3_directory_access_point" "example" {
  bucket = "bucket--zoneid--x-s3"
  name = "example--zoneid--xa-s3"
  account_id = "123456789012"
}
```

### S3 Access Point for a directory bucket with Scope configuration
```
resource "aws_s3_directory_access_point" "example_local_zone" {
    bucket = "bucket--zoneid--x-s3"
    name = "example--zoneid--xa-s3"
    account_id = "123456789012"
    
    scope {
      permissions = ["GetObject", "PutObject"]
      prefixes = ["myobject1.csv", "myobject2*"]
    }
}
```

## Argument Reference
This resource supports the following arguments:
* `name` - (Required) The name you want to assign to this access point. The access point name must consist of a base name that you provide and suffix that includes the ZoneID (AWS Local Zone) of your bucket location, followed by `--xa-s3`. Use the [`aws_s3_access_point`](s3_access_point.html) resource to manage access points for general purpose buckets.

* `bucket` - (Required) The directory bucket that you want to associate this access point with. The name must be in the format `[bucket_name]--[zoneid]--x-s3`. Use the [`aws_s3_bucket`](s3_bucket.html) resource to manage general purpose buckets.

* `account_id` - (Required) The AWS account ID for the account that owns the specified access point.

* `bucket_account_id` (Optional) - The AWS account ID associated with the directory bucket associated with this access point. For same account access point when your bucket and access point belong to the same account owner, the BucketAccountId is not required. For cross-account access point when your bucket and access point are not in the same account, the BucketAccountId is required.

* `policy` - (Optional) Valid JSON document that specifies the policy that you want to apply to this access point. Removing `policy` from your configuration or setting `policy` to null or an empty string (i.e., `policy = ""`) _will not_ delete the policy since it could have been set by `aws_s3control_directory_access_point_policy`. To remove the `policy`, set it to `"{}"` (an empty JSON document).

* `public_access_block_configuration` - (Optional) Block Public Access for directory buckets is turned on by  default and cannot be changed. For more information, see [AWS Documentation](https://docs.aws.amazon.com/AmazonS3/latest/userguide/access-control-block-public-access.html).

* `vpc_configuration` - (Optional) If you include this field, Amazon S3 restricts access to this access point to requests from the specified virtual private cloud (VPC).

* `scope` - (Optional). With access points for directory buckets, you can use the access point scope to restrict access to specific prefixes, API actions, or a combination of both. You can specify any amount of prefixes, but the total length of characters of all prefixes must be less than 256 bytes.

    NOTE: To delete your current `scope`, you must set your scope to `{permissions=[] prefixes=[]}`. A scope set to null _will not_ delete the scope since it could have been set by `aws_s3control_directory_access_point_scope`.


### vpc_configuration Configuration Block
The following arguments are required if you use `vpc_configuration`:
* `vpc_id` - (Required) Amazon S3 restricts access to this access point to requests from the specified virtual private cloud (VPC).


### scope Configuration block
The following arguments are optional:

* `permissions` – (Optional) You can specify a list of API operations as permissions for the access point.

* `prefixes` – (Optional) You can specify a list of prefixes, but the total length of characters of all prefixes must be less than 256 bytes. 

* For more information on access point scope, see [AWS Documentation](https://docs.aws.amazon.com/AmazonS3/latest/userguide/access-points-directory-buckets-manage-scope.html).


## Attribute Reference
This resource exports the following attributes in addition to the arguments above:
* `alias` - The access point alias for the directory bucket. In directory buckets, the access point name and alias are the same value, and can be used interchangably. 
* `arn` - The access point ARN for the directory bucket. In directory buckets, ARN can only be used in access point IAM resource policies and cannot be used for object operations.
* `endpoint` - A list of the AWS service endpoints for the access point.
* `id` - The access point name and AWS account ID separated by a colon (`:`). 
* `network_origin` - Indicates if the access point allows access from the public Internet or restricts to a single VPC. Valid values include `VPC` and `Internet`.


## Import
In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import this resource using `name` and `account_id` separated by a colon (`:`).
For example:

```terraform
import {
  to = aws_s3_directory_access_point.example
  id = "example--zoneid--xa-s3:123456789012"
}
```


**Using `terraform import` to import.**, import access point for directory using the `name` and `account_id` separated by a colon (`:`). For example:

```console
% terraform import aws_s3_directory_access_point.example example--zoneid--xa-s3:123456789012
```