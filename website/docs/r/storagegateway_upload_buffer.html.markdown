---
subcategory: "Storage Gateway"
layout: "aws"
page_title: "AWS: aws_storagegateway_upload_buffer"
description: |-
  Manages an AWS Storage Gateway upload buffer
---

# Resource: aws_storagegateway_upload_buffer

Manages an AWS Storage Gateway upload buffer.

~> **NOTE:** The Storage Gateway API provides no method to remove an upload buffer disk. Destroying this Terraform resource does not perform any Storage Gateway actions.

## Example Usage

### Cached and VTL Gateway Type

```terraform
data "aws_storagegateway_local_disk" "test" {
  disk_node   = aws_volume_attachment.test.device_name
  gateway_arn = aws_storagegateway_gateway.test.arn
}

resource "aws_storagegateway_upload_buffer" "test" {
  disk_path   = data.aws_storagegateway_local_disk.test.disk_path
  gateway_arn = aws_storagegateway_gateway.test.arn
}
```

### Stored Gateway Type

```terraform
data "aws_storagegateway_local_disk" "test" {
  disk_node   = aws_volume_attachment.test.device_name
  gateway_arn = aws_storagegateway_gateway.test.arn
}

resource "aws_storagegateway_upload_buffer" "example" {
  disk_id     = data.aws_storagegateway_local_disk.example.id
  gateway_arn = aws_storagegateway_gateway.example.arn
}
```

## Argument Reference

This resource supports the following arguments:

* `disk_id` - (Optional) Local disk identifier. For example, `pci-0000:03:00.0-scsi-0:0:0:0`.
* `disk_path` - (Optional) Local disk path. For example, `/dev/nvme1n1`.
* `gateway_arn` - (Required) The Amazon Resource Name (ARN) of the gateway.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Combined gateway Amazon Resource Name (ARN) and local disk identifier.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_storagegateway_upload_buffer` using the gateway Amazon Resource Name (ARN) and local disk identifier separated with a colon (`:`). For example:

```terraform
import {
  to = aws_storagegateway_upload_buffer.example
  id = "arn:aws:storagegateway:us-east-1:123456789012:gateway/sgw-12345678:pci-0000:03:00.0-scsi-0:0:0:0"
}
```

Using `terraform import`, import `aws_storagegateway_upload_buffer` using the gateway Amazon Resource Name (ARN) and local disk identifier separated with a colon (`:`). For example:

```console
% terraform import aws_storagegateway_upload_buffer.example arn:aws:storagegateway:us-east-1:123456789012:gateway/sgw-12345678:pci-0000:03:00.0-scsi-0:0:0:0
```
