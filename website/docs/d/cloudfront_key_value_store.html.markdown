---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_key_value_store"
description: |-
  Terraform data source for managing an AWS CloudFront Key Value Store.
---
# Data Source: aws_cloudfront_key_value_store

Terraform data source for managing an AWS CloudFront Key Value Store.

## Example Usage

### Basic Usage

```terraform
data "aws_cloudfront_key_value_store" "example" {
  name = "example_key_value_store"
}

output "key_value_store_arn" {
  value = data.aws_cloudfront_key_value_store.example.arn
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Unique name for your CloudFront KeyValueStore.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) identifying your CloudFront KeyValueStore.
* `comment` - Comment.
* `id` - A unique identifier for the KeyValueStore. Same as `name`.
* `last_modified_time` - The last modified time of the KeyValueStore.
