---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_link_association"
description: |-
  Associates a link to a device.
---

# Resource: aws_networkmanager_link_association

Associates a link to a device.
A device can be associated to multiple links and a link can be associated to multiple devices.
The device and link must be in the same global network and the same site.

## Example Usage

```terraform
resource "aws_networkmanager_link_association" "example" {
  global_network_id = aws_networkmanager_global_network.example.id
  link_id           = aws_networkmanager_link.example.id
  device_id         = aws_networkmanager_device.example.id
}
```

## Argument Reference

This resource supports the following arguments:

* `device_id` - (Required) The ID of the device.
* `global_network_id` - (Required) The ID of the global network.
* `link_id` - (Required) The ID of the link.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_networkmanager_link_association` using the global network ID, link ID and device ID. For example:

```terraform
import {
  to = aws_networkmanager_link_association.example
  id = "global-network-0d47f6t230mz46dy4,link-444555aaabbb11223,device-07f6fd08867abc123"
}
```

Using `terraform import`, import `aws_networkmanager_link_association` using the global network ID, link ID and device ID. For example:

```console
% terraform import aws_networkmanager_link_association.example global-network-0d47f6t230mz46dy4,link-444555aaabbb11223,device-07f6fd08867abc123
```
