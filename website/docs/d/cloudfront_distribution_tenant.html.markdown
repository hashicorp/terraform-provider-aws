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

Exactly one of the following arguments must be specified for the data source:

* `id` (optional) - Identifier for the distribution tenant. For example: `EDFDVBD632BHDS5`.
* `domain` (optional) - An associated domain of the distribution tenant.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `domains` - List of domains for the distribution tenant.

* `arn` - ARN (Amazon Resource Name) for the distribution tenant.

* `status` - Current status of the distribution tenant. `Deployed` if the
    distribution tenant's information is fully propagated throughout the Amazon
    CloudFront system.

* `last_modified_time` - Date and time the distribution tenant was last modified.

* `distribution_id` - The ID of the CloudFront distribution the tenant is associated with.

* `etag` - Current version of the distribution tenant's information. For example:
    `E2QWRUHAPOMQZL`.

* `enabled` - Whether the distribution tenant is enabled.

* `connection_group_id` - The CloudFront connection group the tenant is associated with.
