---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_origin_access_identities"
description: |-
  Use this data source to retrieve information about a set of Amazon CloudFront origin access identities.
---

# Data Source: aws_cloudfront_origin_access_identities

Use this data source to get ARNs, ids and S3 canonical user IDs of Amazon CloudFront origin access identities.

## Example Usage

### All origin access identities in the account

```terraform
data "aws_cloudfront_origin_access_identities" "example" {}
```

### Origin access identities filtered by comment/name

Origin access identities whose comments are `example-comment1`, `example-comment2`

```terraform
data "aws_cloudfront_origin_access_identities" "example" {
  comments = ["example-comment1", "example-comment2"]
}
```

## Argument Reference

* `comments` (Optional) - Filter origin access identities by comment.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:


* `iam_arns` - Set of ARNs of the matched origin access identities.
* `ids` - Set of ids of the matched origin access identities.
* `s3_canonical_user_ids` - Set of S3 canonical user IDs of the matched origin access identities.
