---
subcategory: "CloudFront KeyValueStore"
layout: "aws"
page_title: "AWS: aws_cloudfrontkeyvaluestore_key"
description: |-
  Terraform resource for managing an AWS CloudFront KeyValueStore Key.
---

# Resource: aws_cloudfrontkeyvaluestore_key

Terraform resource for managing an AWS CloudFront KeyValueStore Key.

## Example Usage

### Basic Usage

```terraform
resource "aws_cloudfront_key_value_store" "example" {
  name    = "ExampleKeyValueStore"
  comment = "This is an example key value store"
}

resource "aws_cloudfrontkeyvaluestore_key" "example" {
  key_value_store_arn = aws_cloudfront_key_value_store.example.arn
  key                 = "Test Key"
  value               = "Test Value"
}
```

## Argument Reference

The following arguments are required:

* `key` - (Required) Key to put.
* `key_value_store_arn` - (Required) Amazon Resource Name (ARN) of the Key Value Store.
* `value` - (Required) Value to put.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Combination of attributes separated by a `,` to create a unique id: `key_value_store_arn`,`key`
* `total_size_in_bytes` - Total size of the Key Value Store in bytes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudFront KeyValueStore Key using the `example_id_arg`. For example:

```terraform
import {
  to = aws_cloudfrontkeyvaluestore_key.example
  id = "arn:aws:cloudfront::111111111111:key-value-store/8562g61f-caba-2845-9d99-b97diwae5d3c,someKey"
}
```

Using `terraform import`, import CloudFront KeyValueStore Key using the `id`. For example:

```console
% terraform import aws_cloudfrontkeyvaluestore_key.example arn:aws:cloudfront::111111111111:key-value-store/8562g61f-caba-2845-9d99-b97diwae5d3c,someKey
```
