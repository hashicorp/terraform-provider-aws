---
subcategory: "Storage Gateway"
layout: "aws"
page_title: "AWS: aws_storagegateway_local_disk"
description: |-
  Retrieve information about a Storage Gateway local disk
---

# Data Source: aws_storagegateway_local_disk

Retrieve information about a Storage Gateway local disk. The disk identifier is useful for adding the disk as a cache or upload buffer to a gateway.

## Example Usage

```terraform
data "aws_storagegateway_local_disk" "test" {
  disk_path   = aws_volume_attachment.test.device_name
  gateway_arn = aws_storagegateway_gateway.test.arn
}
```

## Argument Reference

* `gateway_arn` - (Required) ARN of the gateway.
* `disk_node` - (Optional) Device node of the local disk to retrieve. For example, `/dev/sdb`.
* `disk_path` - (Optional) Device path of the local disk to retrieve. For example, `/dev/xvdb` or `/dev/nvme1n1`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `disk_id` - Disk identifierE.g., `pci-0000:03:00.0-scsi-0:0:0:0`
* `id` - Disk identifierE.g., `pci-0000:03:00.0-scsi-0:0:0:0`
