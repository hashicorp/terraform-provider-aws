---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_key_group"
description: |-
  Terraform data source for managing an AWS CloudFront Key Group.
---

# Data Source: aws_cloudfront_key_group

Terraform data source for managing an AWS CloudFront Key Group.

## Example Usage

### Basic Usage

```terraform
data "aws_cloudfront_key_group" "example" {
  name = "example"
}
```

## Argument Reference

The following arguments are required:

* `id` - (Required) ID of the key group. For example: EDFDVBD632BHDS5. Either `id` or `name` must be provided.
* `name` - (Required) Name of the key group. Either `name` or `id` must be provided.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `etag` - The identifier for this version of the key group.
* `comment` - The comment to describe the key group.
* `items` - The list of the identifiers of the public keys in the key group.
