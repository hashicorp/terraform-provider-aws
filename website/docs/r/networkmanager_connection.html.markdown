---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_connection"
description: |-
  Manages a Network Manager Connection.
---

# Resource: aws_networkmanager_connection

Manages a Network Manager Connection.

Use this resource to create a connection between two devices in your global network.

## Example Usage

```terraform
resource "aws_networkmanager_connection" "example" {
  global_network_id   = aws_networkmanager_global_network.example.id
  device_id           = aws_networkmanager_device.example1.id
  connected_device_id = aws_networkmanager_device.example2.id
}
```

## Argument Reference

The following arguments are required:

* `connected_device_id` - (Required) ID of the second device in the connection.
* `device_id` - (Required) ID of the first device in the connection.
* `global_network_id` - (Required) ID of the global network.

The following arguments are optional:

* `connected_link_id` - (Optional) ID of the link for the second device.
* `description` - (Optional) Description of the connection.
* `link_id` - (Optional) ID of the link for the first device.
* `tags` - (Optional) Key-value tags for the connection. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the connection.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `delete` - (Default `10m`)
* `update` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_networkmanager_connection` using the connection ARN. For example:

```terraform
import {
  to = aws_networkmanager_connection.example
  id = "arn:aws:networkmanager::123456789012:device/global-network-0d47f6t230mz46dy4/connection-07f6fd08867abc123"
}
```

Using `terraform import`, import `aws_networkmanager_connection` using the connection ARN. For example:

```console
% terraform import aws_networkmanager_connection.example arn:aws:networkmanager::123456789012:device/global-network-0d47f6t230mz46dy4/connection-07f6fd08867abc123
```
