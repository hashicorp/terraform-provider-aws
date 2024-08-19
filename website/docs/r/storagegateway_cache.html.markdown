---
subcategory: "Storage Gateway"
layout: "aws"
page_title: "AWS: aws_storagegateway_cache"
description: |-
  Manages an AWS Storage Gateway cache
---

# Resource: aws_storagegateway_cache

Manages an AWS Storage Gateway cache.

~> **NOTE:** The Storage Gateway API provides no method to remove a cache disk. Destroying this Terraform resource does not perform any Storage Gateway actions.

## Example Usage

```terraform
resource "aws_storagegateway_cache" "example" {
  disk_id     = data.aws_storagegateway_local_disk.example.id
  gateway_arn = aws_storagegateway_gateway.example.arn
}
```

## Argument Reference

This resource supports the following arguments:

* `disk_id` - (Required) Local disk identifier. For example, `pci-0000:03:00.0-scsi-0:0:0:0`.
* `gateway_arn` - (Required) The Amazon Resource Name (ARN) of the gateway.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Combined gateway Amazon Resource Name (ARN) and local disk identifier.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_storagegateway_cache` using the gateway Amazon Resource Name (ARN) and local disk identifier separated with a colon (`:`). For example:

```terraform
import {
  to = aws_storagegateway_cache.example
  id = "arn:aws:storagegateway:us-east-1:123456789012:gateway/sgw-12345678:pci-0000:03:00.0-scsi-0:0:0:0"
}
```

Using `terraform import`, import `aws_storagegateway_cache` using the gateway Amazon Resource Name (ARN) and local disk identifier separated with a colon (`:`). For example:

```console
% terraform import aws_storagegateway_cache.example arn:aws:storagegateway:us-east-1:123456789012:gateway/sgw-12345678:pci-0000:03:00.0-scsi-0:0:0:0
```
