---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_connection"
description: |-
  Creates a connection between two devices.
---

# Resource: aws_networkmanager_connection

Creates a connection between two devices.
The devices can be a physical or virtual appliance that connects to a third-party appliance in a VPC, or a physical appliance that connects to another physical appliance in an on-premises network.

## Example Usage

```terraform
resource "aws_networkmanager_connection" "example" {
  global_network_id   = aws_networkmanager_global_network.example.id
  device_id           = aws_networkmanager_device.example1.id
  connected_device_id = aws_networkmanager_device.example2.id
}
```

## Argument Reference

The following arguments are supported:

* `connected_device_id` - (Required) The ID of the second device in the connection.
* `connected_link_id` - (Optional) The ID of the link for the second device.
* `description` - (Optional) A description of the connection.
* `device_id` - (Required) The ID of the first device in the connection.
* `global_network_id` - (Required) The ID of the global network.
* `link_id` - (Optional) The ID of the link for the first device.
* `tags` - (Optional) Key-value tags for the connection. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the connection.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

`aws_networkmanager_connection` can be imported using the connection ARN, e.g.

```
$ terraform import aws_networkmanager_connection.example arn:aws:networkmanager::123456789012:device/global-network-0d47f6t230mz46dy4/connection-07f6fd08867abc123
```
