---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_origin_access_identities"
description: |-
  Use this data source to retrieve information about a set of Amazon CloudFront origin access identities.
---

# Data Source: aws_cloudfront_origin_access_identities

Use this data source to get ARNs, ids, S3 canonical user IDs of Amazon CloudFront origin access identities.

## Example Usage

### All origin access identities in the account

```terraform
data "aws_cloudfront_origin_access_identities" "example" {

}
```

### Origin access identities filtered by comment/name

Origin access identities whose comments are `example-comment1`, `example-comment2`

```terraform
data "aws_cloudfront_origin_access_identities" "example" {
  	filter {
		name = "comment"
		values = ["example-comment1","example-comment2"]
	}
}
```

## Argument Reference

* `filter` (Optional) - Object that determines whether to filter based on certain parameter and return only a set of CloudFront origin access identities. See [Filter Config](#filter-config) for more information.

### Filter Config

* `name` (Required) - Name of the parameter to filter. Valid value is `comment`.
* `values` (Required) - A list of existing origin access identities comments.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:


* `arns` - Set of ARNs of the matched origin access identities.
* `ids` - Set of ids of the matched origin access identities.
* `s3_canonical_user_ids` - Set of S3 canonical user IDs of the matched origin access identities.