---
subcategory: "CloudFront KeyValueStore"
layout: "aws"
page_title: "AWS: aws_cloudfrontkeyvaluestore_keys_exclusive"
description: |-
  Terraform resource for maintaining exclusive management of resource key value pairs defined in an AWS CloudFront KeyValueStore.
---
# Resource: aws_cloudfrontkeyvaluestore_keys_exclusive

Terraform resource for maintaining exclusive management of resource key value pairs defined in an AWS CloudFront KeyValueStore.

!> This resource takes exclusive ownership over key value pairs defined in a KeyValueStore. This includes removal of key value pairs which are not explicitly configured. To prevent persistent drift, ensure any [`aws_cloudfrontkeyvaluestore_key`](./cloudfrontkeyvaluestore_key.html.markdown) resources managed alongside this resource have an equivalent `resource_key_value_pair` argument.

~> Destruction of this resource means Terraform will no longer manage reconciliation of the configured key value pairs. It __will not__ delete the configured key value pairs from the KeyValueStore.

## Example Usage

### Basic Usage

```terraform
resource "aws_cloudfront_key_value_store" "example" {
  name    = "ExampleKeyValueStore"
  comment = "This is an example key value store"
}

resource "aws_cloudfrontkeyvaluestore_keys_exclusive" "example" {
  key_value_store_arn = aws_cloudfront_key_value_store.example.arn

  resource_key_value_pair {
    key   = "Test Key"
    value = "Test Value"
  }
}
```

### Disallow Key Value Pairs

To automatically remove any configured key value pairs, omit a `resource_key_value_pair` block.

~> This will not __prevent__ key value pairs from being defined in a KeyValueStore via Terraform (or any other interface). This resource enables bringing key value pairs into a configured state, however, this reconciliation happens only when `apply` is proactively run.

```terraform
resource "aws_cloudfrontkeyvaluestore_keys_exclusive" "example" {
  key_value_store_arn = aws_cloudfront_key_value_store.example.arn
}
```

## Argument Reference

The following arguments are required:

* `key_value_store_arn` - (Required) Amazon Resource Name (ARN) of the Key Value Store.

The following arguments are optional:

* `max_batch_size` - (Optional) Maximum resource key values pairs that will update in a single API request. AWS has a default quota of 50 keys or a 3 MB payload, whichever is reached first. Defaults to `50`.
* `resource_key_value_pair` - (Optional) A list of all resource key value pairs associated with the KeyValueStore.
See [`resource_key_value_pair`](#resource_key_value_pair) below.

### `resource_key_value_pair`

The following arguments are required:

* `key` - (Required) Key to put.
* `value` - (Required) Value to put.

## Attribute Reference

This resource exports no additional attributes.

* `total_size_in_bytes` - Total size of the Key Value Store in bytes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AWS CloudFront KeyValueStore Key Value Pairs using the `key_value_store_arn`. For example:

```terraform
import {
  to = aws_cloudfrontkeyvaluestore_keys_exclusive.example
  id = "arn:aws:cloudfront::111111111111:key-value-store/8562g61f-caba-2845-9d99-b97diwae5d3c"
}
```

Using `terraform import`, import AWS CloudFront KeyValueStore Key Value Pairs using the `key_value_store_arn`. For example:

```console
% terraform import aws_cloudfrontkeyvaluestore_keys_exclusive.example arn:aws:cloudfront::111111111111:key-value-store/8562g61f-caba-2845-9d99-b97diwae5d3c
```
