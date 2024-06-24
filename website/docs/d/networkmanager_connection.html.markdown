---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_connection"
description: |-
  Retrieve information about a connection.
---

# Data Source:  aws_networkmanager_connection

Retrieve information about a connection.

## Example Usage

```terraform
data "aws_networkmanager_connection" "example" {
  global_network_id = var.global_network_id
  connection_id     = var.connection_id
}
```

## Argument Reference

* `global_network_id` - (Required) ID of the Global Network of the connection to retrieve.
* `connection_id` - (Required) ID of the specific connection to retrieve.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the connection.
* `connected_device_id` - ID of the second device in the connection.
* `connected_link_id` - ID of the link for the second device.
* `description` - Description of the connection.
* `device_id` - ID of the first device in the connection.
* `link_id` - ID of the link for the first device.
* `tags` - Key-value tags for the connection.
