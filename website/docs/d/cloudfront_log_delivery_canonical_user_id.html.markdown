---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_log_delivery_canonical_user_id"
description: |-
  Provides the canonical user ID of the AWS `awslogsdelivery` account for CloudFront bucket logging.
---

# Data Source: aws_cloudfront_log_delivery_canonical_user_id

The CloudFront Log Delivery Canonical User ID data source allows access to the [canonical user ID](http://docs.aws.amazon.com/general/latest/gr/acct-identifiers.html) of the AWS `awslogsdelivery` account for CloudFront bucket logging.
See the [Amazon CloudFront Developer Guide](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/AccessLogs.html) for more information.

## Example Usage

```terraform
data "aws_cloudfront_log_delivery_canonical_user_id" "example" {}

resource "aws_s3_bucket" "example" {
  bucket = "example"
}

resource "aws_s3_bucket_acl" "example" {
  bucket = aws_s3_bucket.example.id

  access_control_policy {
    grant {
      grantee {
        id   = data.aws_cloudfront_log_delivery_canonical_user_id.example.id
        type = "CanonicalUser"
      }
      permission = "FULL_CONTROL"
    }
  }
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region you'd like the zone for. By default, fetches the current region.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - Canonical user ID for the AWS `awslogsdelivery` account in the region.
