---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_key_group"
description: |-
  Provides a CloudFront key group.
---

# Resource: aws_cloudfront_key_group

## Example Usage

The following example below creates a CloudFront key group.

```terraform
resource "aws_cloudfront_public_key" "example" {
  comment     = "example public key"
  encoded_key = file("public_key.pem")
  name        = "example-key"
}

resource "aws_cloudfront_key_group" "example" {
  comment = "example key group"
  items   = [aws_cloudfront_public_key.example.id]
  name    = "example-key-group"
}
```

## Argument Reference

This resource supports the following arguments:

* `comment` - (Optional) A comment to describe the key group..
* `items` - (Required) A list of the identifiers of the public keys in the key group.
* `name` - (Required) A name to identify the key group.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `etag` - The identifier for this version of the key group.
* `id` - The identifier for the key group.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudFront Key Group using the `id`. For example:

```terraform
import {
  to = aws_cloudfront_key_group.example
  id = "4b4f2r1c-315d-5c2e-f093-216t50jed10f"
}
```

Using `terraform import`, import CloudFront Key Group using the `id`. For example:

```console
% terraform import aws_cloudfront_key_group.example 4b4f2r1c-315d-5c2e-f093-216t50jed10f
```
