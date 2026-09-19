---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_invalidation"
description: |-
  Terraform resource for managing an AWS CloudFront Invalidation.
---

# Resource: aws_cloudfront_invalidation

Terraform resource for managing an AWS CloudFront Invalidations.

## Example Usage

### Basic Usage

```terraform
resource "aws_cloudfront_invalidation" "example" {
  distribution_id = aws_cloudfront_distribution.example.id
  paths = [
    "/*",
  ]
}
```

## Argument Reference

The following arguments are required:

* `distribution_id` - (Required) CloudFront Distribution ID.
* `paths` - (Required) List of paths to invalidate.
* `triggers` (Optional) Map of arbitrary keys and values that, when changed, will trigger a re-invocation. To force a re-invocation without changing these keys/values, use the [`terraform taint` command](https://www.terraform.io/docs/commands/taint.html).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ID of the invalidation.
* `status` - Invalidation status.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)
