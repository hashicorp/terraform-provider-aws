---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_distribution_tenant"
description: |-
  Provides a CloudFront distribution tenant data source.
---

# Data Source: aws_cloudfront_distribution_tenant

Use this data source to retrieve information about a CloudFront distribution tenant.

## Example Usage

```terraform
data "aws_cloudfront_distribution_tenant" "test" {
  id = "EDFDVBD632BHDS5"
}
```

## Argument Reference

This data source supports the following arguments:

* `id` (Optional) - Identifier for the distribution tenant. For example: `EDFDVBD632BHDS5`. Exactly one of `id` or `domain` must be specified.
* `domain` (Optional) - An associated domain of the distribution tenant. Exactly one of `id` or `domain` must be specified.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `domains` - List of domains for the distribution tenant.
* `arn` - ARN (Amazon Resource Name) for the distribution tenant.
* `status` - Current status of the distribution tenant. `Deployed` if the
    distribution tenant's information is fully propagated throughout the Amazon
    CloudFront system.
* `distribution_id` - The ID of the CloudFront distribution the tenant is associated with.
* `etag` - Current version of the distribution tenant's information. For example:
    `E2QWRUHAPOMQZL`.
* `enabled` - Whether the distribution tenant is enabled.
* `connection_group_id` - The CloudFront connection group the tenant is associated with.
