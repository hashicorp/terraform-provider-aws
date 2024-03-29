---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_origin_access_identity"
description: |-
  Use this data source to retrieve information for an Amazon CloudFront origin access identity.
---

# Data Source: aws_cloudfront_origin_access_identity

Use this data source to retrieve information for an Amazon CloudFront origin access identity.

## Example Usage

The following example below creates a CloudFront origin access identity.

```terraform
data "aws_cloudfront_origin_access_identity" "example" {
  id = "E1ZAKK699EOLAL"
}
```

## Argument Reference

* `id` (Required) -  The identifier for the origin access identity. For example: `E1ZAKK699EOLAL`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `caller_reference` - Internal value used by CloudFront to allow future
   updates to the origin access identity.
* `cloudfront_access_identity_path` - A shortcut to the full path for the
   origin access identity to use in CloudFront, see below.
* `comment` - An optional comment for the origin access identity.
* `etag` - Current version of the origin access identity's information.
   For example: `E2QWRUHAPOMQZL`.
* `iam_arn` - Pre-generated ARN for use in S3 bucket policies (see below).
   Example: `arn:aws:iam::cloudfront:user/CloudFront Origin Access Identity
   E2QWRUHAPOMQZL`.
* `s3_canonical_user_id` - The Amazon S3 canonical user ID for the origin
   access identity, which you use when giving the origin access identity read
   permission to an object in Amazon S3.
