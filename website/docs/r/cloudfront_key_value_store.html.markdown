---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_key_value_store"
description: |-
  Terraform resource for managing an AWS CloudFront Key Value Store.
---

# Resource: aws_cloudfront_key_value_store

Terraform resource for managing an AWS CloudFront Key Value Store.

## Example Usage

### Basic Usage

```terraform
resource "aws_cloudfront_key_value_store" "example" {
  name    = "ExampleKeyValueStore"
  comment = "This is an example key value store"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Unique name for your CloudFront KeyValueStore.

The following arguments are optional:

* `comment` - (Optional) Comment.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) identifying your CloudFront KeyValueStore.
* `id` - A unique identifier for the KeyValueStore. Same as `name`.
* `etag` - ETag hash of the KeyValueStore.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudFront Key Value Store using the `name`. For example:

```terraform
import {
  to = aws_cloudfront_key_value_store.example
  id = "example_store"
}
```

Using `terraform import`, import CloudFront Key Value Store using the `name`. For example:

```console
% terraform import aws_cloudfront_key_value_store.example example_store
```
