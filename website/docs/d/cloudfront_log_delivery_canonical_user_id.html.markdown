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
data "aws_canonical_user_id" "current" {}

data "aws_cloudfront_log_delivery_canonical_user_id" "example" {}

resource "aws_s3_bucket" "example" {
  bucket = "example"
}

resource "aws_s3_bucket_ownership_controls" "example" {
  bucket = aws_s3_bucket.example.id
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
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
    owner {
      id = data.aws_canonical_user_id.current.id
    }
  }
  depends_on = [aws_s3_bucket_ownership_controls.example]
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Name of the Region whose canonical user ID is desired. Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - Canonical user ID for the AWS `awslogsdelivery` account in the Region.
