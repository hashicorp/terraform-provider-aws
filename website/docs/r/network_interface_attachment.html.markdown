---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_network_interface_attachment"
description: |-
  Attach an Elastic network interface (ENI) resource with EC2 instance.
---

# Resource: aws_network_interface_attachment

Attach an Elastic network interface (ENI) resource with EC2 instance.

## Example Usage

```terraform
resource "aws_network_interface_attachment" "test" {
  instance_id          = aws_instance.test.id
  network_interface_id = aws_network_interface.test.id
  device_index         = 0
}
```

## Argument Reference

This resource supports the following arguments:

* `instance_id` - (Required) Instance ID to attach.
* `network_interface_id` - (Required) ENI ID to attach.
* `device_index` - (Required) Network interface index (int).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `instance_id` - Instance ID.
* `network_interface_id` - Network interface ID.
* `attachment_id` - The ENI Attachment ID.
* `status` - The status of the Network Interface Attachment.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Elastic network interface (ENI) Attachments using its Attachment ID. For example:

```terraform
import {
  to = aws_network_interface_attachment.secondary_nic
  id = "eni-attach-0a33842b4ec347c4c"
}
```

Using `terraform import`, import Elastic network interface (ENI) Attachments using its Attachment ID. For example:

```console
% terraform import aws_network_interface_attachment.secondary_nic eni-attach-0a33842b4ec347c4c
```
