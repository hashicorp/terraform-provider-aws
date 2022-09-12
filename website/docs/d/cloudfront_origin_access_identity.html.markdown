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
  id = "EDFDVBD632BHDS5"
}
```

## Argument Reference

* `id` (Required) -  The identifier for the distribution. For example: `EDFDVBD632BHDS5`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

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

## Using With CloudFront

Normally, when referencing an origin access identity in CloudFront, you need to
prefix the ID with the `origin-access-identity/cloudfront/` special path.
The `cloudfront_access_identity_path` allows this to be circumvented.
The below snippet demonstrates use with the `s3_origin_config` structure for the
[`aws_cloudfront_distribution`][3] resource:

```terraform
resource "aws_cloudfront_distribution" "example" {
  # ... other configuration ...

  origin {
    s3_origin_config {
      origin_access_identity = aws_cloudfront_origin_access_identity.example.cloudfront_access_identity_path
    }
  }
}
```

### Updating your bucket policy

Note that the AWS API may translate the `s3_canonical_user_id` `CanonicalUser`
principal into an `AWS` IAM ARN principal when supplied in an
[`aws_s3_bucket`][4] bucket policy, causing spurious diffs in Terraform. If
you see this behaviour, use the `iam_arn` instead:

```terraform
data "aws_iam_policy_document" "s3_policy" {
  statement {
    actions   = ["s3:GetObject"]
    resources = ["${aws_s3_bucket.example.arn}/*"]

    principals {
      type        = "AWS"
      identifiers = [aws_cloudfront_origin_access_identity.example.iam_arn]
    }
  }
}

resource "aws_s3_bucket_policy" "example" {
  bucket = aws_s3_bucket.example.id
  policy = data.aws_iam_policy_document.s3_policy.json
}
```

[1]: http://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/Introduction.html
[2]: http://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/private-content-restricting-access-to-s3.html
[3]: /docs/providers/aws/r/cloudfront_distribution.html
[4]: /docs/providers/aws/r/s3_bucket.html


## Import

Cloudfront Origin Access Identities can be imported using the `id`, e.g.,

```
$ terraform import aws_cloudfront_origin_access_identity.origin_access E74FTE3AEXAMPLE
```
