---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_connection_function"
description: |-
  Manages an AWS CloudFront Connection Function.
---

# Resource: aws_cloudfront_connection_function

Manages an AWS CloudFront Connection Function.

## Example Usage

### Basic Usage

```terraform
resource "aws_cloudfront_connection_function" "example" {
  name    = "example-connection-function"
  runtime = "cloudfront-js-2.0"
  comment = "Example connection function"
  code    = "function handler(event) { return event.request; }"
}
```

### With Publish Enabled

```terraform
resource "aws_cloudfront_connection_function" "example" {
  name    = "example-connection-function"
  runtime = "cloudfront-js-2.0"
  comment = "Example connection function with publish enabled"
  code    = "function handler(event) { return event.request; }"
  publish = true
}
```

### With Key Value Store Associations

```terraform
resource "aws_cloudfront_key_value_store" "example" {
  name    = "example-kvs"
  comment = "Example key value store"
}

resource "aws_cloudfront_connection_function" "example" {
  name    = "example-connection-function"
  runtime = "cloudfront-js-2.0"
  comment = "Example connection function with KVS"
  code    = "function handler(event) { return event.request; }"

  key_value_store_associations {
    items {
      key_value_store_arn = aws_cloudfront_key_value_store.example.arn
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `code` - (Required) Code for the connection function. Maximum length is 40960 characters.
* `name` - (Required) Name for the connection function. Must be 1-64 characters and can contain letters, numbers, hyphens, and underscores.
* `runtime` - (Required) Runtime environment for the function. Valid values are `cloudfront-js-1.0` and `cloudfront-js-2.0`.

The following arguments are optional:

* `comment` - (Optional) Comment to describe the function.
* `key_value_store_associations` - (Optional) Configuration for key value store associations. See [Key Value Store Associations](#key-value-store-associations) below. AWS limits associations to one key value store per function.
* `publish` - (Optional) Whether to publish the function to the `LIVE` stage after creation or update. Defaults to `false`.

### Key Value Store Associations

* `items` - (Optional) List of key value store associations.
    * `key_value_store_arn` - (Required) ARN of the key value store.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Connection Function.
* `created_time` - Date and time when the connection function was created.
* `etag` - ETag of the connection function.
* `id` - ID of the connection function.
* `last_modified_time` - Date and time when the connection function was last modified.
* `live_stage_etag` - ETag of the function's LIVE stage. Will be empty if the function has not been published.
* `location` - Location of the connection function.
* `status` - Status of the connection function.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `update` - (Default `10m`)
* `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudFront Connection Function using the function ID. For example:

```terraform
import {
  to = aws_cloudfront_connection_function.example
  id = "E1PA6795UKMFR9"
}
```

Using `terraform import`, import CloudFront Connection Function using the function ID. For example:

```console
% terraform import aws_cloudfront_connection_function.example E1PA6795UKMFR9
```
