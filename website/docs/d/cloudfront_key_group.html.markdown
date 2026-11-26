---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_key_group"
description: |-
  Use this data source to retrieve information about a CloudFront key group.
---

# Data Source: aws_cloudfront_key_group

Use this data source to retrieve information about a CloudFront key group.

## Example Usage

### By Name

```terraform
data "aws_cloudfront_key_group" "example" {
  name = "example-key-group"
}
```

### By ID

```terraform
data "aws_cloudfront_key_group" "example" {
  id = "c8b6e3a1-7d2f-4c9e-a1b0-123456789abc"
}
```

## Argument Reference

This data source supports the following arguments:

* `id` - (Optional) Identifier of the key group. Exactly one of `id` or `name` must be specified.
* `name` - (Optional) Name of the key group. Exactly one of `id` or `name` must be specified.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `comment` - Comment describing the key group.
* `etag` - Identifier for this version of the key group.
* `items` - Set of identifiers of the public keys in the key group.
* `last_modified_time` - Date and time when the key group was last modified, in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8).
