---
layout: "aws"
page_title: "AWS: cloudfront_distribution"
sidebar_current: "docs-aws-datasource-cloudfront-distribution"
description: |-
  Provides metadata for a CloudFront web distribution resource.
---

# aws_cloudfront_distribution

This data source allows access to useful information related to a specific Cloudfront distribution.

For information about CloudFront distributions, see the
[Amazon CloudFront Developer Guide][1]. For specific information about reading
CloudFront web distribution information, see the [GET Distribution][2] page in the Amazon
CloudFront API Reference.

## Example Usage

The following example below reads data from a CloudFront distribution with an S3 origin.

```hcl
data "aws_cloudfront_distribution" "s3_distribution" {
  id = "EDFDVBD632BHDS5"
}
```

## Argument Reference

The following arguments are supported:

* `id` - (Required) The ID of the distribution

## Attribute Reference

The following attributes are exported:

* `id` - The identifier for the distribution. For example: `EDFDVBD632BHDS5`.

* `arn` - The ARN (Amazon Resource Name) for the distribution. For example: arn:aws:cloudfront::123456789012:distribution/EDFDVBD632BHDS5, where 123456789012 is your AWS account ID.

* `status` - The current status of the distribution. `Deployed` if the
  distribution's information is fully propagated throughout the Amazon
  CloudFront system.

* `active_trusted_signers` - The key pair IDs that CloudFront is aware of for
  each trusted signer, if the distribution is set up to serve private content
  with signed URLs.

* `domain_name` - The domain name corresponding to the distribution. For
  example: `d604721fxaaqy9.cloudfront.net`.

* `last_modified_time` - The date and time the distribution was last modified.

* `in_progress_validation_batches` - The number of validation batches
  currently in progress.

* `etag` - The current version of the distribution's information. For example:
  `E2QWRUHAPOMQZL`.

[1]: http://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/Introduction.html
[2]: https://docs.aws.amazon.com/cloudfront/latest/APIReference/API_GetDistribution.html
