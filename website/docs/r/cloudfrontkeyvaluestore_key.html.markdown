---
subcategory: "CloudFront KeyValueStore"
layout: "aws"
page_title: "AWS: aws_cloudfrontkeyvaluestore_key"
description: |-
  Terraform resource for managing an AWS CloudFront KeyValueStore Key.
---

# Resource: aws_cloudfrontkeyvaluestore_key

Terraform resource for managing an AWS CloudFront KeyValueStore Key.

!> This resource manages individual key value pairs in a KeyValueStore. This can lead to high costs associated with accessing the CloudFront KeyValueStore API when performing terraform operations with many key value pairs defined. For large key value stores, consider the [`aws_cloudfrontkeyvaluestore_keys_exclusive`](./cloudfrontkeyvaluestore_keys_exclusive.html.markdown) resource to minimize the number of API calls made to the CloudFront KeyValueStore API.

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

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_cloudfrontkeyvaluestore_key.example
  identity = {
    key_value_store_arn = "arn:aws:cloudfront::111111111111:key-value-store/8562g61f-caba-2845-9d99-b97diwae5d3c"
    key                 = "someKey"
  }
}

resource "aws_cloudfrontkeyvaluestore_key" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `key_value_store_arn` (String) ARN of the CloudFront Key Value Store.
* `key` (String) Key name.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudFront KeyValueStore Key using the `key_value_store_arn` and 'key' separated by `,`. For example:

```terraform
import {
  to = aws_cloudfrontkeyvaluestore_key.example
  id = "arn:aws:cloudfront::111111111111:key-value-store/8562g61f-caba-2845-9d99-b97diwae5d3c,someKey"
}
```

Using `terraform import`, import CloudFront KeyValueStore Key using the `key_value_store_arn` and 'key' separated by `,`. For example:

```console
% terraform import aws_cloudfrontkeyvaluestore_key.example arn:aws:cloudfront::111111111111:key-value-store/8562g61f-caba-2845-9d99-b97diwae5d3c,someKey
```
