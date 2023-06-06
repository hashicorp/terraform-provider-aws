---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_connections"
description: |-
  Retrieve information about connections.
---

# Data Source: aws_networkmanager_connections

Retrieve information about connections.

## Example Usage

```terraform
data "aws_networkmanager_connections" "example" {
  global_network_id = var.global_network_id

  tags = {
    Env = "test"
  }
}
```

## Argument Reference

* `device_id` - (Optional) ID of the device of the connections to retrieve.
* `global_network_id` - (Required) ID of the Global Network of the connections to retrieve.
* `tags` - (Optional) Restricts the list to the connections with these tags.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `ids` - IDs of the connections.
