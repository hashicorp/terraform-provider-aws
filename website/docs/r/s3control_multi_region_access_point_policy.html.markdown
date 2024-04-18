---
subcategory: "S3 Control"
layout: "aws"
page_title: "AWS: aws_s3control_multi_region_access_point_policy"
description: |-
  Provides a resource to manage an S3 Multi-Region Access Point access control policy.
---

# Resource: aws_s3control_multi_region_access_point_policy

Provides a resource to manage an S3 Multi-Region Access Point access control policy.

## Example Usage

### Basic Example

```terraform
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_s3_bucket" "foo_bucket" {
  bucket = "example-bucket-foo"
}

resource "aws_s3control_multi_region_access_point" "example" {
  details {
    name = "example"

    region {
      bucket = aws_s3_bucket.foo_bucket.id
    }
  }
}

resource "aws_s3control_multi_region_access_point_policy" "example" {
  details {
    name = element(split(":", aws_s3control_multi_region_access_point.example.id), 1)
    policy = jsonencode({
      "Version" : "2012-10-17",
      "Statement" : [
        {
          "Sid" : "Example",
          "Effect" : "Allow",
          "Principal" : {
            "AWS" : data.aws_caller_identity.current.account_id
          },
          "Action" : ["s3:GetObject", "s3:PutObject"],
          "Resource" : "arn:${data.aws_partition.current.partition}:s3::${data.aws_caller_identity.current.account_id}:accesspoint/${aws_s3control_multi_region_access_point.example.alias}/object/*"
        }
      ]
    })
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `account_id` - (Optional) The AWS account ID for the owner of the Multi-Region Access Point. Defaults to automatically determined account ID of the Terraform AWS provider.
* `details` - (Required) A configuration block containing details about the policy for the Multi-Region Access Point. See [Details Configuration Block](#details-configuration) below for more details

### Details Configuration

The `details` block supports the following:

* `name` - (Required) The name of the Multi-Region Access Point.
* `policy` - (Required) A valid JSON document that specifies the policy that you want to associate with this Multi-Region Access Point. Once applied, the policy can be edited, but not deleted. For more information, see the documentation on [Multi-Region Access Point Permissions](https://docs.aws.amazon.com/AmazonS3/latest/userguide/MultiRegionAccessPointPermissions.html).

-> **NOTE:** When you update the `policy`, the update is first listed as the proposed policy. After the update is finished and all Regions have been updated, the proposed policy is listed as the established policy. If both policies have the same version number, the proposed policy is the established policy.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `established` - The last established policy for the Multi-Region Access Point.
* `id` - The AWS account ID and access point name separated by a colon (`:`).
* `proposed` - The proposed policy for the Multi-Region Access Point.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `15m`)
* `update` - (Default `15m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Multi-Region Access Point Policies using the `account_id` and `name` of the Multi-Region Access Point separated by a colon (`:`). For example:

```terraform
import {
  to = aws_s3control_multi_region_access_point_policy.example
  id = "123456789012:example"
}
```

Using `terraform import`, import Multi-Region Access Point Policies using the `account_id` and `name` of the Multi-Region Access Point separated by a colon (`:`). For example:

```console
% terraform import aws_s3control_multi_region_access_point_policy.example 123456789012:example
```
