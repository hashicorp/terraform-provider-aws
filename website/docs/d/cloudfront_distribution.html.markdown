---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_distribution"
description: |-
  Provides a CloudFront web distribution data source.
---

# Data Source: aws_cloudfront_distribution

Use this data source to retrieve information about a CloudFront distribution.

## Example Usage

```terraform
data "aws_cloudfront_distribution" "test" {
  id = "EDFDVBD632BHDS5"
}
```

## Argument Reference

* `id` - Identifier for the distribution. For example: `EDFDVBD632BHDS5`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - Identifier for the distribution. For example: `EDFDVBD632BHDS5`.

* `aliases` - List that contains information about CNAMEs (alternate domain names), if any, for this distribution.

* `arn` - ARN (Amazon Resource Name) for the distribution. For example: arn:aws:cloudfront::123456789012:distribution/EDFDVBD632BHDS5, where 123456789012 is your AWS account ID.

* `status` - Current status of the distribution. `Deployed` if the
    distribution's information is fully propagated throughout the Amazon
    CloudFront system.

* `domain_name` - Domain name corresponding to the distribution. For
    example: `d604721fxaaqy9.cloudfront.net`.

* `last_modified_time` - Date and time the distribution was last modified.

* `in_progress_validation_batches` - The number of invalidation batches
    currently in progress.

* `etag` - Current version of the distribution's information. For example:
    `E2QWRUHAPOMQZL`.

* `hosted_zone_id` - CloudFront Route 53 zone ID that can be used to
     route an [Alias Resource Record Set][7] to. This attribute is simply an
     alias for the zone ID `Z2FDTNDATAQYW2`.
* `web_acl_id` AWS WAF web ACL associated with this distribution.
