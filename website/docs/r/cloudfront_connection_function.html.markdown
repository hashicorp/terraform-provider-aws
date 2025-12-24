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

### With Tags

```terraform
resource "aws_cloudfront_connection_function" "example" {
  name    = "example-connection-function"
  runtime = "cloudfront-js-2.0"
  comment = "Example connection function with tags"
  code    = "function handler(event) { return event.request; }"

  tags = {
    Environment = "production"
    Team        = "web"
  }
}
```

## Argument Reference

The following arguments are required:

* `code` - (Required) Code for the connection function. Maximum length is 40960 characters.
* `name` - (Required) Name for the connection function. Must be 1-64 characters and can contain letters, numbers, hyphens, and underscores. Changing this forces a new resource to be created.
* `runtime` - (Required) Runtime environment for the function. Valid values are `cloudfront-js-1.0` and `cloudfront-js-2.0`.

The following arguments are optional:

* `comment` - (Optional) Comment to describe the function.
* `key_value_store_associations` - (Optional) Configuration block for key value store associations. See [`key_value_store_associations`](#key_value_store_associations) below.
* `publish` - (Optional) Whether to publish the function to the `LIVE` stage after creation or update. Defaults to `false`.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### key_value_store_associations

* `items` - (Required) List of key value store associations.
    * `key_value_store_arn` - (Required) ARN of the key value store.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the connection function.
* `etag` - ETag of the connection function.
* `id` - ID of the connection function.
* `live_stage_etag` - ETag of the function's LIVE stage. Will be empty if the function has not been published.
* `status` - Status of the connection function.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

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
